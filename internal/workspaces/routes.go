package workspaces

import (
	"web-based-dev-platform-backend/internal/config"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/gorm"
)

func RegisterWorkspaces(r chi.Router, db *pgxpool.Pool, cfg *config.Config) {
	r.Post("/projects/{projectID}/workspaces", createWorkspace(db))
	r.Post("/workspaces/{id}/start", startWorkspace(db, cfg))
	r.Post("/workspaces/{id}/stop", stopWorkspace(db, cfg))
	r.Delete("/workspaces/{id}", deleteWorkspace(db, cfg))
	r.Get("/workspaces/{id}/status", workspaceStatus(db))
}

func RegisterRoutes(r chi.Router, db *gorm.DB) {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	r.Post("/workspaces", handler.CreateWorkspace)
}
