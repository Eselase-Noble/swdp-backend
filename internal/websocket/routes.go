package websocket

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterWebSockets(r chi.Router, db *pgxpool.Pool, cfg Config) {
	r.Get("/ws/logs/{workspaceID}", LogsHandler(cfg))
	r.Get("/ws/terminal/{workspaceID}", TerminalHandler(cfg))
}
