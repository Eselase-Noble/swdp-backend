package projects

import (
	"encoding/json"
	"net/http"
	"time"

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

//Create a new project
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

//List all the projects assigned to the logged user
func listProjects(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value("userId").(string)

		rows, err := db.Query(
			r.Context(),
			`SELECT projectId, projectName, description FROM projects 
				WHERE createdBy = $1 OR projectId IN (SELECT projectId FROM project_members  WHERE userId = $1)`, userId,
		)

		if err != nil {
			http.Error(w, "Error listing projects", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var projects []map[string]string
		for rows.Next() {
			var projectId, projectName, description string
			rows.Scan(&projectId, &projectName, &description)
			projects = append(projects, map[string]string{
				"projectId":   projectId,
				"projectName": projectName,
				"description": description,
			})
		}

		json.NewEncoder(w).Encode(projects)
	}
}
