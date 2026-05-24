package websocket

import (
	"context"
	"io"
	"log"
	"net/http"
	"web-based-dev-platform-backend/internal/runtime"

	"github.com/coder/websocket"
	"github.com/go-chi/chi/v5"
)

func LogsHandler(rt runtime.Runtime) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workspaceID := chi.URLParam(r, "workspaceID")

		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			OriginPatterns: []string{"*"},
		})
		if err != nil {
			log.Println("logs: websocket accept:", err)
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "")

		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		reader, err := rt.Logs(ctx, workspaceID)
		if err != nil {
			log.Println("logs: get workspace logs:", err)
			conn.Close(websocket.StatusInternalError, "failed to get workspace logs")
			return
		}
		defer reader.Close()

		buf := make([]byte, 4096)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				n, err := reader.Read(buf)
				if err != nil {
					if err != io.EOF {
						log.Println("logs: read:", err)
					}
					return
				}
				if err := conn.Write(ctx, websocket.MessageText, buf[:n]); err != nil {
					return
				}
			}
		}
	}
}
