package websocket

import (
	"web-based-dev-platform-backend/internal/runtime"

	"github.com/go-chi/chi/v5"
)

func RegisterWebSockets(r chi.Router, rt runtime.Runtime) {
	r.Get("/ws/logs/{workspaceID}", LogsHandler(rt))
	r.Get("/ws/terminal/{workspaceID}", TerminalHandler(rt))
}
