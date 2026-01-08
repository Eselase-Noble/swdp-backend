package workspaces

import (
	"web-based-dev-platform-backend/internal/config"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterWorkspaces(r chi.Router, db *pgxpool.Pool, cfg *config.Config) {
	r.Post("/projects/{projectID}/workspaces", createWorkspace(db))
	r.Post("/workspaces/{id}/start", startWorkspace(db, cfg))
	r.Post("/workspaces/{id}/stop", stopWorkspace(db, cfg))
	r.Delete("/workspaces/{id}", deleteWorkspace(db, cfg))
	r.Get("/workspaces/{id}/status", workspaceStatus(db))
}

//func assertWorkSpaceOwner(ctx context.Context, db *pgxpool.Pool, wsID, userid string) error {
//	var ok bool
//	err := db.QueryRow(ctx, `SELECT  EXISTS ( SELECT 1 FROM workspaces WHERE id = $1 AND userId = $2`, wsID, userid).Scan(&ok)
//	if err != nil || !ok {
//		return errors.New("workspace not owned by this user")
//	}
//	return nil
//}
//
//func createWorkspace(db *pgxpool.Pool) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		projectID := chi.URLParam(r, "projectID")
//		userID := r.Context().Value("userId").(string)
//
//		id := uuid.New()
//
//		_, err := db.Exec(
//			r.Context(),
//			"INSERT INTO workspaces(id, project_id, user_id, status) VALUES ($1,$2,$3,'created')",
//			id, projectID, userID,
//		)
//		if err != nil {
//			http.Error(w, "db error", http.StatusInternalServerError)
//			return
//		}
//
//		w.WriteHeader(http.StatusCreated)
//	}
//}
//
//func startWorkspace(db *pgxpool.Pool, cfg *config.Config) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		wsID := chi.URLParam(r, "id")
//		userID := r.Context().Value("userId").(string)
//
//		if err := assertWorkSpaceOwner(r.Context(), db, wsID, userID); err != nil {
//			http.Error(w, "forbidden", 403)
//			return
//		}
//
//		var status string
//		db.QueryRow(r.Context(), "SELECT status FROM workspaces WHERE id=$1", wsID).Scan(&status)
//		if status == "running" {
//			w.WriteHeader(http.StatusConflict)
//			return
//		}
//
//		err := execution.StartContainer(
//			cfg.DockerClient,
//			"ws-"+wsID,
//			"node:20",
//			"/data/workspaces/"+wsID,
//		)
//		if err != nil {
//			http.Error(w, err.Error(), http.StatusInternalServerError)
//			return
//		}
//
//		db.Exec(r.Context(),
//			"UPDATE workspaces SET status='running' WHERE id=$1", wsID)
//
//		w.WriteHeader(http.StatusOK)
//	}
//}
//
//func stopWorkspace(db *pgxpool.Pool, cfg *config.Config) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		wsID := chi.URLParam(r, "id")
//		userID := r.Context().Value("userID").(string)
//
//		if err := assertWorkSpaceOwner(r.Context(), db, wsID, userID); err != nil {
//			http.Error(w, "forbidden", 403)
//			return
//		}
//
//		execution.StopContainer(cfg.DockerClient, "ws-"+wsID)
//
//		_, err := db.Exec(r.Context(),
//			"UPDATE workspaces SET status='stopped' WHERE id=$1", wsID)
//		if err != nil {
//			return
//		}
//
//		w.WriteHeader(http.StatusOK)
//	}
//}
//
//func deleteWorkspace(db *pgxpool.Pool, cfg *config.Config) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		wsID := chi.URLParam(r, "id")
//		userId := r.Context().Value("userId").(string)
//
//		if err := assertWorkSpaceOwner(r.Context(), db, wsID, userId); err != nil {
//			http.Error(w, "forbidden", http.StatusForbidden)
//			return
//		}
//
//		execution.StopContainer(cfg.DockerClient, "ws-"+wsID)
//
//		_, err := db.Exec(r.Context(), `DELETE FROM workspaces WHERE id = $1`, wsID)
//		if err != nil {
//			return
//		}
//		w.WriteHeader(http.StatusNoContent)
//	}
//}
//
//func workspaceStatus(db *pgxpool.Pool) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		wsID := chi.URLParam(r, "id")
//		userID := r.Context().Value("userID").(string)
//
//		if err := assertWorkSpaceOwner(r.Context(), db, wsID, userID); err != nil {
//			http.Error(w, "forbidden", 403)
//			return
//		}
//
//		var status string
//		err := db.QueryRow(r.Context(), "SELECT status FROM workspaces WHERE id=$1", wsID).Scan(&status)
//		if err != nil {
//			http.Error(w, "not found", 404)
//			return
//		}
//
//		json.NewEncoder(w).Encode(map[string]string{
//			"status": status,
//		})
//	}
//}
