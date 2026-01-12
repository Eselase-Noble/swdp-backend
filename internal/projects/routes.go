package projects

import (
	"encoding/json"
	"net/http"
	"time"
	"web-based-dev-platform-backend/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type createProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func RegisterProjects(r chi.Router, db *pgxpool.Pool) {
	r.Post("/projects/create-project", createProject(db))
	r.Get("/projects/all", listProjects(db))
}

// Create a new project
func createProject(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value("userId").(string)

		var req createProjectRequest

		_ = json.NewDecoder(r.Body).Decode(&req)

		id := uuid.New()

		_, err := db.Exec(
			r.Context(),
			`INSERT INTO projects(projectId, projectName, description,createdBy, createdAt, updatedBy, updatedAt)
				 VALUES($1, $2, $3, $4, $5, $6, $7)`,
			id, req.Name, req.Description, userId, time.Now(), userId, time.Now(),
		)
		if err != nil {
			http.Error(w, "Error creating a new project", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

// List all the projects assigned to the logged user
//func listProjects(db *pgxpool.Pool) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		//userId := r.Context().Value("userId").(string)
//
//		rows, err := db.Query(
//			r.Context(),
//			`SELECT projectId, projectName, description
//             FROM projects
//             WHERE projectId IN (SELECT projectId FROM project_members) OR createdBy IS NOT NULL`,
//		)
//
//		if err != nil {
//			http.Error(w, "Error listing projects", http.StatusInternalServerError)
//			return
//		}
//		defer rows.Close()
//
//		var projects []map[string]string
//		for rows.Next() {
//			var projectId, projectName, description string
//			rows.Scan(&projectId, &projectName, &description)
//			projects = append(projects, map[string]string{
//				"projectId":   projectId,
//				"projectName": projectName,
//				"description": description,
//			})
//		}
//
//		json.NewEncoder(w).Encode(projects)
//	}
//}

// ListProjects godoc
// @Summary      List projects
// @Description  Get all projects for the authenticated user
// @Tags         Projects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} models.Project
// @Failure      401 {object} map[string]string
// @Router       /projects [get]
func listProjects(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		user := middleware.GetUser(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		rows, err := db.Query(r.Context(), `SELECT project_id, project_name, owner_id FROM projects`)
		if err != nil {
			http.Error(w, "Error listing projects", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var projects []map[string]string
		for rows.Next() {
			var project_id, project_name, owner_id string
			if err := rows.Scan(&project_id, &project_name, &owner_id); err != nil {
				http.Error(w, "Error scanning project row", http.StatusInternalServerError)
				return
			}
			projects = append(projects, map[string]string{
				"projectId":   project_id,
				"projectName": project_name,
				"owner":       owner_id,
			})
		}

		w.Header().Set("Content-Type", "application/json")

		if len(projects) == 0 {
			// No projects found
			json.NewEncoder(w).Encode(map[string]string{
				"message": "No project available at the moment",
			})
			return
		}

		// Projects found
		json.NewEncoder(w).Encode(projects)
	}
}
