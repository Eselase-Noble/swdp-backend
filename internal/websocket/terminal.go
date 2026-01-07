package websocket

import (
	"context"
	"io"
	"log"
	"net/http"

	"github.com/coder/websocket"
	"github.com/go-chi/chi/v5"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func TerminalHandler(docker *client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		workspaceID := chi.URLParam(r, "workspaceID")
		containerName := "ws-" + workspaceID

		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "")

		ctx := context.Background()

		// Create exec instance
		execConfig := types.ExecConfig{
			Cmd:          []string{"/bin/sh"},
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          true,
		}

		execResp, err := docker.ContainerExecCreate(ctx, containerName, execConfig)
		if err != nil {
			log.Println("exec create error:", err)
			return
		}

		// Attach to exec instance
		attachResp, err := docker.ContainerExecAttach(ctx, execResp.ID, types.ExecStartCheck{
			Detach: false,
			Tty:    true,
		})
		if err != nil {
			log.Println("exec attach error:", err)
			return
		}
		defer attachResp.Close()

		// Create a cancellable context for goroutine synchronization
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		// WS → container stdin
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

		// container stdout/stderr → WS
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

		// Wait for context cancellation (when any goroutine finishes)
		<-ctx.Done()
	}
}
