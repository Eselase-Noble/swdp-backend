package runtime

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
)

const workspaceImageTag = "swdp-workspace:latest"

// workspaceDockerfileContent is embedded at compile-time. It installs every
// runtime SWDP supports: Python 3, Node 20 + npx, Go 1.22, Java 17 (OpenJDK).
const workspaceDockerfileContent = `FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive

# Base tools + Python 3 + Java 17
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl git ca-certificates build-essential \
    python3 python3-pip \
    openjdk-17-jdk-headless \
 && rm -rf /var/lib/apt/lists/*

# Node.js 20 LTS (includes npm + npx)
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
 && apt-get install -y nodejs \
 && rm -rf /var/lib/apt/lists/*

# Go 1.22
RUN ARCH=$(dpkg --print-architecture | sed 's/x86_64/amd64/') && \
    curl -fsSL "https://go.dev/dl/go1.22.4.linux-${ARCH}.tar.gz" | tar -C /usr/local -xzf -
ENV PATH=$PATH:/usr/local/go/bin

# ts-node for TypeScript execution
RUN npm install -g ts-node typescript --quiet

WORKDIR /workspace
CMD ["sleep", "infinity"]
`

// DockerRuntime runs each workspace as a Docker container backed by a named volume.
// Container naming: ws-<id>    Volume naming: vol-<id>
type DockerRuntime struct {
	client *client.Client
}

func NewDockerRuntime(cli *client.Client) *DockerRuntime {
	return &DockerRuntime{client: cli}
}

// ensureImage guarantees cfg.Image exists locally. For swdp-workspace:latest it
// builds from the embedded Dockerfile; for other images it falls back to pull.
func (d *DockerRuntime) ensureImage(ctx context.Context, image string) error {
	if _, _, err := d.client.ImageInspectWithRaw(ctx, image); err == nil {
		return nil // already present
	}

	if image == workspaceImageTag {
		return d.buildWorkspaceImage(ctx)
	}

	// Generic image — pull from registry
	reader, err := d.client.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("docker pull %s: %w", image, err)
	}
	_, _ = io.Copy(io.Discard, reader)
	return reader.Close()
}

// buildWorkspaceImage builds swdp-workspace:latest from the embedded Dockerfile.
// Takes ~3-5 min on first run; all subsequent creates are instant.
func (d *DockerRuntime) buildWorkspaceImage(ctx context.Context) error {
	log.Println("runtime: building swdp-workspace:latest (first-time setup, ~3-5 min)…")

	content := []byte(workspaceDockerfileContent)
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	if err := tw.WriteHeader(&tar.Header{
		Name:     "Dockerfile",
		Mode:     0o644,
		Size:     int64(len(content)),
		Typeflag: tar.TypeReg,
	}); err != nil {
		return fmt.Errorf("build context: %w", err)
	}
	if _, err := tw.Write(content); err != nil {
		return fmt.Errorf("build context write: %w", err)
	}
	_ = tw.Close()

	resp, err := d.client.ImageBuild(ctx, &buf, types.ImageBuildOptions{
		Tags:       []string{workspaceImageTag},
		Dockerfile: "Dockerfile",
		Remove:     true,
	})
	if err != nil {
		return fmt.Errorf("docker build workspace image: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body) // drain — blocks until build completes

	log.Println("runtime: swdp-workspace:latest built successfully")
	return nil
}

func (d *DockerRuntime) Create(ctx context.Context, cfg WorkspaceConfig) error {
	if err := d.ensureImage(ctx, cfg.Image); err != nil {
		return err
	}

	volName := volumeName(cfg.ID)
	if _, err := d.client.VolumeCreate(ctx, volume.CreateOptions{Name: volName}); err != nil {
		return fmt.Errorf("docker create volume: %w", err)
	}

	containerCfg := &container.Config{
		Image:      cfg.Image,
		WorkingDir: "/workspace",
		Tty:        true,
		Cmd:        []string{"sleep", "infinity"},
	}
	hostCfg := &container.HostConfig{
		Binds:       []string{volName + ":/workspace"},
		NetworkMode: "none",
		SecurityOpt: []string{"no-new-privileges"},
		Resources: container.Resources{
			Memory:   cfg.MemoryMB * 1024 * 1024,
			NanoCPUs: cfg.CPUs * 1_000_000_000,
		},
	}

	if _, err := d.client.ContainerCreate(ctx, containerCfg, hostCfg, nil, nil, containerName(cfg.ID)); err != nil {
		if !errdefs.IsConflict(err) {
			return fmt.Errorf("docker create container: %w", err)
		}
	}
	return nil
}

func (d *DockerRuntime) Start(ctx context.Context, id string) error {
	return d.client.ContainerStart(ctx, containerName(id), types.ContainerStartOptions{})
}

func (d *DockerRuntime) Stop(ctx context.Context, id string) error {
	timeout := 10
	return d.client.ContainerStop(ctx, containerName(id), container.StopOptions{Timeout: &timeout})
}

func (d *DockerRuntime) Delete(ctx context.Context, id string) error {
	name := containerName(id)
	_ = d.client.ContainerStop(ctx, name, container.StopOptions{})
	if err := d.client.ContainerRemove(ctx, name, types.ContainerRemoveOptions{Force: true}); err != nil {
		return fmt.Errorf("docker remove container: %w", err)
	}
	if err := d.client.VolumeRemove(ctx, volumeName(id), false); err != nil {
		return fmt.Errorf("docker remove volume: %w", err)
	}
	return nil
}

func (d *DockerRuntime) Status(ctx context.Context, id string) (Status, error) {
	info, err := d.client.ContainerInspect(ctx, containerName(id))
	if err != nil {
		return StatusUnknown, err
	}
	switch info.State.Status {
	case "running":
		return StatusRunning, nil
	case "exited", "dead":
		return StatusStopped, nil
	default:
		return StatusCreated, nil
	}
}

func (d *DockerRuntime) Attach(ctx context.Context, id string) (io.ReadWriteCloser, error) {
	execID, err := d.client.ContainerExecCreate(ctx, containerName(id), types.ExecConfig{
		Cmd:          []string{"/bin/bash"},
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("docker exec create: %w", err)
	}
	resp, err := d.client.ContainerExecAttach(ctx, execID.ID, types.ExecStartCheck{Tty: true})
	if err != nil {
		return nil, fmt.Errorf("docker exec attach: %w", err)
	}
	return &dockerConn{resp}, nil
}

func (d *DockerRuntime) Logs(ctx context.Context, id string) (io.ReadCloser, error) {
	return d.client.ContainerLogs(ctx, containerName(id), types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       "100",
	})
}

func containerName(id string) string { return "ws-" + id }
func volumeName(id string) string    { return "vol-" + id }

type dockerConn struct{ resp types.HijackedResponse }

func (c *dockerConn) Read(p []byte) (int, error)  { return c.resp.Reader.Read(p) }
func (c *dockerConn) Write(p []byte) (int, error) { return c.resp.Conn.Write(p) }
func (c *dockerConn) Close() error                { c.resp.Close(); return nil }
