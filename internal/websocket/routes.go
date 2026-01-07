package websocket

import (
	"github.com/docker/docker/client"
	"github.com/go-chi/chi/v5"
)

func RegisterWebSockets(r chi.Router, dockerClient *client.Client) {
	r.Get("/ws/logs/{workspaceID}", LogsHandler(dockerClient))
	r.Get("/ws/terminal/{workspaceID}", TerminalHandler(dockerClient))
}
