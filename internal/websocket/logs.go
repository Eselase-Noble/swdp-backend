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

func LogsHandler(dockerClient *client.Client) http.HandlerFunc {
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

		// Get container logs
		reader, err := dockerClient.ContainerLogs(ctx, containerName, types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true,
			Timestamps: false,
			Tail:       "100",
		})
		if err != nil {
			log.Println("container logs error:", err)
			conn.Close(websocket.StatusInternalError, "Failed to get container logs")
			return
		}
		defer reader.Close()

		// Create a cancellable context
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		// Stream logs to WebSocket
		go func() {
			defer cancel()
			buf := make([]byte, 4096)
			for {
				select {
				case <-ctx.Done():
					return
				default:
					n, err := reader.Read(buf)
					if err != nil {
						if err != io.EOF {
							log.Println("logs read error:", err)
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
