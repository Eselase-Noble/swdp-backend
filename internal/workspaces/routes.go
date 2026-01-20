package workspaces

import "github.com/go-chi/chi/v5"

func Routes(r chi.Router, h *Handler) {
	r.Post("/create-workpace", h.CreateWorkspace)
	r.Post("/{id}/start", h.StartWorkspace)
	r.Post("/{id}/stop", h.StopWorkspace)
	r.Get("/{id}/status", h.WorkspaceStatus)
	r.Delete("/{id}/delete", h.DeleteWorkspace)
}
