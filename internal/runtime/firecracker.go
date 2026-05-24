package runtime

// Firecracker MicroVM runtime.
//
// Each workspace runs as a real MicroVM (not a container). VMs start in ~125 ms,
// use ~5 MB of overhead, and are fully isolated at the kernel level.
//
// ─── Host prerequisites ────────────────────────────────────────────────────────
//
//  1. KVM must be available on the host:
//       ls /dev/kvm          # file must exist
//       sudo chmod 666 /dev/kvm
//
//  2. Firecracker binary (v1.x):
//       curl -Lo /usr/local/bin/firecracker \
//         https://github.com/firecracker-microvm/firecracker/releases/download/v1.9.1/firecracker-v1.9.1-x86_64
//       chmod +x /usr/local/bin/firecracker
//
//  3. Kernel binary (uncompressed vmlinux):
//       # Download a pre-built one from the Firecracker team:
//       curl -Lo /opt/firecracker/vmlinux \
//         https://s3.amazonaws.com/spec.ccfc.min/img/quickstart_guide/x86_64/kernels/vmlinux.bin
//
//  4. Base rootfs image (ext4):
//       # Ubuntu 22.04 minimal rootfs with socat installed.
//       # See /opt/firecracker/README.md after running make-rootfs.sh.
//       # The rootfs MUST run this on boot (e.g. in /etc/rc.local):
//         socat VSOCK-LISTEN:9999,fork,reuseaddr EXEC:/bin/bash,pty,rawer &
//       # This exposes a shell on vsock port 9999 for the terminal WebSocket.
//
// ─── How it works ──────────────────────────────────────────────────────────────
//
//  Create  → copy base.ext4  → /var/swdp/workspaces/<id>/rootfs.ext4
//  Start   → launch Firecracker process with that rootfs, store Machine handle
//  Stop    → graceful ACPI shutdown (rootfs data is preserved on disk)
//  Delete  → stop + rm -rf workspace directory
//  Attach  → connect via vsock Unix socket, send "CONNECT 9999\n" → shell I/O
//  Logs    → read Firecracker process log file

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	firecracker "github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	"github.com/sirupsen/logrus"
)

// FirecrackerConfig holds the host-level paths the runtime needs.
type FirecrackerConfig struct {
	KernelPath   string // absolute path to vmlinux binary
	RootfsBase   string // absolute path to the base ext4 rootfs template
	WorkspaceDir string // directory under which per-workspace dirs are created
}

type firecrackerRuntime struct {
	cfg      FirecrackerConfig
	machines sync.Map // workspace id -> *firecracker.Machine
}

// workspaceMeta is persisted to disk at Create time so Start can read it
// after a server restart without needing the original WorkspaceConfig.
type workspaceMeta struct {
	CPUs     int64 `json:"cpus"`
	MemoryMB int64 `json:"memory_mb"`
}

func NewFirecrackerRuntime(cfg FirecrackerConfig) (*firecrackerRuntime, error) {
	if err := os.MkdirAll(cfg.WorkspaceDir, 0700); err != nil {
		return nil, fmt.Errorf("create workspace root dir: %w", err)
	}
	return &firecrackerRuntime{cfg: cfg}, nil
}

func (f *firecrackerRuntime) Create(ctx context.Context, cfg WorkspaceConfig) error {
	dir := f.dir(cfg.ID)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("mkdir workspace: %w", err)
	}

	// Each workspace gets its own private copy of the rootfs so changes are
	// isolated between developers on the same project.
	dst := filepath.Join(dir, "rootfs.ext4")
	if err := copyFile(f.cfg.RootfsBase, dst); err != nil {
		return fmt.Errorf("copy rootfs: %w", err)
	}

	meta := workspaceMeta{CPUs: cfg.CPUs, MemoryMB: cfg.MemoryMB}
	b, _ := json.Marshal(meta)
	return os.WriteFile(filepath.Join(dir, "meta.json"), b, 0600)
}

func (f *firecrackerRuntime) Start(ctx context.Context, id string) error {
	if _, running := f.machines.Load(id); running {
		return nil // already running
	}

	meta, err := f.readMeta(id)
	if err != nil {
		return fmt.Errorf("read workspace meta: %w", err)
	}

	dir := f.dir(id)

	// Inline pointer helpers so we don't depend on SDK helper naming across versions.
	pStr := func(s string) *string { return &s }
	pI64 := func(i int64) *int64 { return &i }
	pBool := func(b bool) *bool { return &b }

	machineCfg := firecracker.Config{
		SocketPath:      filepath.Join(dir, "api.sock"),
		LogPath:         filepath.Join(dir, "firecracker.log"),
		LogLevel:        "Info",
		KernelImagePath: f.cfg.KernelPath,
		// console=ttyS0 → serial output goes to Firecracker process stdout
		// reboot=k      → reboot in VM triggers Firecracker exit (not a host reboot)
		KernelArgs: "console=ttyS0 reboot=k panic=1 pci=off",
		Drives: []models.Drive{
			{
				DriveID:      pStr("rootfs"),
				PathOnHost:   pStr(filepath.Join(dir, "rootfs.ext4")),
				IsRootDevice: pBool(true),
				IsReadOnly:   pBool(false),
			},
		},
		MachineCfg: models.MachineConfiguration{
			VcpuCount:  pI64(meta.CPUs),
			MemSizeMib: pI64(meta.MemoryMB),
			HtEnabled:  pBool(false),
		},
		// vsock device: host-side socket = dir/vsock.sock
		// Inside the VM, socat listens on vsock port 9999 and forks a shell.
		VsockDevices: []firecracker.VsockDevice{
			{
				Path: filepath.Join(dir, "vsock.sock"),
				CID:  3,
			},
		},
	}

	// Suppress noisy logrus output; real logs go to LogPath file above.
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	m, err := firecracker.NewMachine(ctx, machineCfg,
		firecracker.WithLogger(logrus.NewEntry(logger)),
	)
	if err != nil {
		return fmt.Errorf("create firecracker machine: %w", err)
	}

	if err := m.Start(ctx); err != nil {
		return fmt.Errorf("start firecracker machine: %w", err)
	}

	f.machines.Store(id, m)
	return nil
}

func (f *firecrackerRuntime) Stop(ctx context.Context, id string) error {
	m, ok := f.loadMachine(id)
	if !ok {
		return nil // already stopped
	}
	defer f.machines.Delete(id)

	stopCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := m.Shutdown(stopCtx); err != nil {
		// Graceful shutdown timed out — force kill the Firecracker process.
		return m.StopVMM()
	}
	return nil
}

func (f *firecrackerRuntime) Delete(ctx context.Context, id string) error {
	_ = f.Stop(ctx, id)
	return os.RemoveAll(f.dir(id))
}

func (f *firecrackerRuntime) Status(ctx context.Context, id string) (Status, error) {
	if _, running := f.machines.Load(id); running {
		return StatusRunning, nil
	}
	if _, err := os.Stat(f.dir(id)); os.IsNotExist(err) {
		return StatusUnknown, fmt.Errorf("workspace %s does not exist", id)
	}
	return StatusStopped, nil
}

// Attach connects to the VM shell over the vsock Unix socket.
//
// Firecracker host-side vsock protocol:
//  1. Dial the Unix socket at dir/vsock.sock
//  2. Send "CONNECT <port>\n"
//  3. Firecracker responds "OK <cid>.<port>\n" — consume that line
//  4. The connection is now a raw bidirectional channel to the guest process
func (f *firecrackerRuntime) Attach(ctx context.Context, id string) (io.ReadWriteCloser, error) {
	vsockPath := filepath.Join(f.dir(id), "vsock.sock")

	conn, err := net.Dial("unix", vsockPath)
	if err != nil {
		return nil, fmt.Errorf("dial vsock socket: %w", err)
	}

	if _, err := fmt.Fprintf(conn, "CONNECT 9999\n"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("vsock CONNECT: %w", err)
	}

	// Read the "OK <cid>.<port>" handshake line before handing the conn to the caller.
	scanner := bufio.NewScanner(conn)
	scanner.Scan()
	if line := scanner.Text(); len(line) < 2 || line[:2] != "OK" {
		conn.Close()
		return nil, fmt.Errorf("vsock handshake failed: %q", line)
	}

	return conn, nil
}

// Logs returns the Firecracker process log for this workspace.
// For VM application output, configure the rootfs to write logs to a file
// and expose them via vsock or a separate log port.
func (f *firecrackerRuntime) Logs(ctx context.Context, id string) (io.ReadCloser, error) {
	logPath := filepath.Join(f.dir(id), "firecracker.log")
	return os.Open(logPath)
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func (f *firecrackerRuntime) dir(id string) string {
	return filepath.Join(f.cfg.WorkspaceDir, id)
}

func (f *firecrackerRuntime) readMeta(id string) (*workspaceMeta, error) {
	b, err := os.ReadFile(filepath.Join(f.dir(id), "meta.json"))
	if err != nil {
		return nil, err
	}
	var m workspaceMeta
	return &m, json.Unmarshal(b, &m)
}

func (f *firecrackerRuntime) loadMachine(id string) (*firecracker.Machine, bool) {
	v, ok := f.machines.Load(id)
	if !ok {
		return nil, false
	}
	return v.(*firecracker.Machine), true
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
