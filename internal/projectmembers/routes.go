package projectmembers

import (
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, handler *Handler) {
	r.Post("/project-members", handler.AddProjectMember)
}
