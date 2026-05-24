package runtime

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

// DockerRuntime runs each workspace as a Docker container backed by a named volume.
// Container naming: ws-<id>
// Volume naming:    vol-<id>
type DockerRuntime struct {
	client *client.Client
}

func NewDockerRuntime(cli *client.Client) *DockerRuntime {
	return &DockerRuntime{client: cli}
}

func (d *DockerRuntime) Create(ctx context.Context, cfg WorkspaceConfig) error {
	volName := volumeName(cfg.ID)

	if _, err := d.client.VolumeCreate(ctx, volume.CreateOptions{Name: volName}); err != nil {
		return fmt.Errorf("docker create volume: %w", err)
	}

	containerCfg := &container.Config{
		Image:      cfg.Image,
		WorkingDir: "/workspace",
		Tty:        true,
		// sleep infinity keeps the container alive so developers can exec into it
		Cmd: []string{"sleep", "infinity"},
	}

	hostCfg := &container.HostConfig{
		Binds:       []string{volName + ":/workspace"},
		NetworkMode: "none", // no outbound internet; protects company code
		SecurityOpt: []string{"no-new-privileges"},
		Resources: container.Resources{
			Memory:   cfg.MemoryMB * 1024 * 1024,
			NanoCPUs: cfg.CPUs * 1_000_000_000,
		},
	}

	if _, err := d.client.ContainerCreate(ctx, containerCfg, hostCfg, nil, nil, containerName(cfg.ID)); err != nil {
		return fmt.Errorf("docker create container: %w", err)
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

// Attach opens an interactive /bin/sh session inside the container.
func (d *DockerRuntime) Attach(ctx context.Context, id string) (io.ReadWriteCloser, error) {
	execID, err := d.client.ContainerExecCreate(ctx, containerName(id), types.ExecConfig{
		Cmd:          []string{"/bin/sh"},
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

// dockerConn wraps Docker's HijackedResponse as io.ReadWriteCloser.
type dockerConn struct {
	resp types.HijackedResponse
}

func (c *dockerConn) Read(p []byte) (int, error)  { return c.resp.Reader.Read(p) }
func (c *dockerConn) Write(p []byte) (int, error) { return c.resp.Conn.Write(p) }
func (c *dockerConn) Close() error                { c.resp.Close(); return nil }
