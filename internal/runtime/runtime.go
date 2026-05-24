package runtime

import (
	"context"
	"io"
)

type Status string

const (
	StatusCreated Status = "created"
	StatusRunning Status = "running"
	StatusStopped Status = "stopped"
	StatusUnknown Status = "unknown"
)

// WorkspaceConfig holds the resource settings for a new workspace.
type WorkspaceConfig struct {
	ID       string // workspace UUID
	Image    string // Docker image name; ignored by Firecracker (uses base rootfs)
	CPUs     int64  // vCPUs to allocate
	MemoryMB int64  // RAM in megabytes
}

// Runtime is the single interface every execution backend must satisfy.
// Swap Docker ↔ Firecracker by changing RUNTIME_TYPE in your env — no other
// code needs to change.
type Runtime interface {
	// Create provisions the workspace's resources (volume + container, or rootfs copy).
	// Called once when the workspace is first made.
	Create(ctx context.Context, cfg WorkspaceConfig) error

	// Start brings an existing workspace online.
	Start(ctx context.Context, id string) error

	// Stop halts the workspace. All data and resources are preserved.
	Stop(ctx context.Context, id string) error

	// Delete removes the workspace and every resource it owns (volume, rootfs, etc.).
	Delete(ctx context.Context, id string) error

	// Status returns the current live state from the runtime (not the DB).
	Status(ctx context.Context, id string) (Status, error)

	// Attach opens a bidirectional shell session (drives the WebSocket terminal).
	Attach(ctx context.Context, id string) (io.ReadWriteCloser, error)

	// Logs returns a streaming reader of workspace output (drives the WebSocket log viewer).
	Logs(ctx context.Context, id string) (io.ReadCloser, error)
}
