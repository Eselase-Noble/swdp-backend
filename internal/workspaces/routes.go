package workspaces

import "github.com/go-chi/chi/v5"

func Routes(r chi.Router, h *Handler) {
	r.Route("/workspaces", func(r chi.Router) {
		r.Post("/", h.CreateWorkspace)
		r.Post("/{id}/start", h.StartWorkspace)
		r.Post("/{id}/stop", h.StopWorkspace)
		r.Get("/{id}/status", h.WorkspaceStatus)
		r.Delete("/{id}", h.DeleteWorkspace)
	})
}
