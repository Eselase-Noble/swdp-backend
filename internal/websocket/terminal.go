package websocket

import (
	"context"
	"io"
	"log"
	"net/http"

	"github.com/coder/websocket"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/go-chi/chi/v5"
)

func TerminalHandler(dockerClient *client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workspaceID := chi.URLParam(r, "workspaceID")
		containerName := "ws-" + workspaceID

		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			log.Println("websocket accept error:", err)
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "")

		ctx := context.Background()

		// Create exec configuration
		execConfig := types.ExecConfig{
			Cmd:          []string{"/bin/sh"},
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          true,
		}

		// Create exec instance in the container
		execResp, err := dockerClient.ContainerExecCreate(ctx, containerName, execConfig)
		if err != nil {
			log.Println("exec create error:", err)
			conn.Close(websocket.StatusInternalError, "Failed to create exec instance")
			return
		}

		// Attach to the exec instance
		attachResp, err := dockerClient.ContainerExecAttach(ctx, execResp.ID, types.ExecStartCheck{
			Detach: false,
			Tty:    true,
		})
		if err != nil {
			log.Println("exec attach error:", err)
			conn.Close(websocket.StatusInternalError, "Failed to attach to exec")
			return
		}
		defer attachResp.Close()

		// Create a cancellable context for goroutine synchronization
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		// WebSocket → container stdin
		go func() {
			defer cancel()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					_, msg, err := conn.Read(ctx)
					if err != nil {
						// Send EOF (Ctrl+D) to terminal when WebSocket closes
						attachResp.Conn.Write([]byte{4})
						return
					}
					_, err = attachResp.Conn.Write(msg)
					if err != nil {
						log.Println("write to container error:", err)
						return
					}
				}
			}
		}()

		// container stdout/stderr → WebSocket
		go func() {
			defer cancel()
			buf := make([]byte, 4096)
			for {
				select {
				case <-ctx.Done():
					return
				default:
					n, err := attachResp.Reader.Read(buf)
					if err != nil {
						if err != io.EOF {
							log.Println("container read error:", err)
						}
						return
					}
					err = conn.Write(ctx, websocket.MessageText, buf[:n])
					if err != nil {
						return
					}
				}
			}
		}()

		// Wait for context cancellation
		<-ctx.Done()
	}
}
