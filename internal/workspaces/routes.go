package workspaces

import "github.com/go-chi/chi/v5"

func Routes(r chi.Router, h *Handler) {
	r.Route("/workspaces", func(r chi.Router) {
		// Workspace lifecycle
		r.Get("/", h.ListWorkspaces)
		r.Post("/", h.CreateWorkspace)
		r.Post("/{id}/start", h.StartWorkspace)
		r.Post("/{id}/stop", h.StopWorkspace)
		r.Get("/{id}/status", h.WorkspaceStatus)
		r.Delete("/{id}", h.DeleteWorkspace)

		// File system API (DockerRuntime only)
		r.Get("/{id}/files", h.ListFiles)
		r.Get("/{id}/files/*", h.ReadFile)
		r.Put("/{id}/files/*", h.WriteFile)
		r.Delete("/{id}/files/*", h.DeleteFile)
		r.Post("/{id}/dirs/*", h.CreateDir)
	})
}
