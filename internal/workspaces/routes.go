package workspaces

import (
	"net/http"
	"web-based-dev-platform-backend/internal/config"
	"web-based-dev-platform-backend/internal/execution"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterWorkspaces(r chi.Router, db *pgxpool.Pool, cfg *config.Config) {
	r.Post("/projects/{projectID}/workspaces", createWorkspace(db))
	r.Post("/workspaces/{id}/start", startWorkspace(db, cfg))
	r.Post("/workspaces/{id}/stop", stopWorkspace(db, cfg))
}

func createWorkspace(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := chi.URLParam(r, "projectID")
		userID := r.Context().Value("userID").(string)

		id := uuid.New()

		_, err := db.Exec(
			r.Context(),
			"INSERT INTO workspaces(id, project_id, user_id, status) VALUES ($1,$2,$3,'created')",
			id, projectID, userID,
		)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func startWorkspace(db *pgxpool.Pool, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wsID := chi.URLParam(r, "id")

		err := execution.StartContainer(
			cfg.DockerClient,
			"ws-"+wsID,
			"node:20",
			"/data/workspaces/"+wsID,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		db.Exec(r.Context(),
			"UPDATE workspaces SET status='running' WHERE id=$1", wsID)

		w.WriteHeader(http.StatusOK)
	}
}

func stopWorkspace(db *pgxpool.Pool, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wsID := chi.URLParam(r, "id")

		execution.StopContainer(cfg.DockerClient, "ws-"+wsID)

		db.Exec(r.Context(),
			"UPDATE workspaces SET status='stopped' WHERE id=$1", wsID)

		w.WriteHeader(http.StatusOK)
	}
}
