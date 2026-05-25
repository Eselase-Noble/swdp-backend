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

func TerminalHandler(rt runtime.Runtime) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workspaceID := chi.URLParam(r, "workspaceID")

		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			// Restrict to your frontend origin in production, e.g. "app.swdp.io"
			OriginPatterns: []string{"*"},
		})
		if err != nil {
			log.Println("terminal: websocket accept:", err)
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "")

		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		rwc, err := rt.Attach(ctx, workspaceID)
		if err != nil {
			log.Println("terminal: attach workspace:", err)
			conn.Close(websocket.StatusInternalError, "failed to attach to workspace")
			return
		}
		defer rwc.Close()

		// WebSocket → workspace stdin
		go func() {
			defer cancel()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					_, msg, err := conn.Read(ctx)
					if err != nil {
						rwc.Write([]byte{4}) // Ctrl+D (EOF) to the shell
						return
					}
					if _, err := rwc.Write(msg); err != nil {
						return
					}
				}
			}
		}()

		// workspace stdout/stderr → WebSocket
		go func() {
			defer cancel()
			buf := make([]byte, 4096)
			for {
				select {
				case <-ctx.Done():
					return
				default:
					n, err := rwc.Read(buf)
					if err != nil {
						if err != io.EOF {
							log.Println("terminal: read from workspace:", err)
						}
						return
					}
					if err := conn.Write(ctx, websocket.MessageBinary, buf[:n]); err != nil {
						return
					}
				}
			}
		}()

		<-ctx.Done()
	}
}
