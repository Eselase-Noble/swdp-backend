package projects

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type Handler struct {
	Service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{Service: service}
}

//
// Create Project
//

// CreateProject godoc
// @Summary      Create project
// @Description  Creates a new project owned by the authenticated user
// @Tags         Projects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        project  body      Project  true  "Project payload"
// @Success      201
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /projects [post]
func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.Context().Value("userId").(string))
	if err != nil {
		http.Error(w, `{"error":"invalid user context"}`, http.StatusUnauthorized)
		return
	}

	var req Project
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	if err := h.Service.CreateProject(r.Context(), userID, req.Name); err != nil {
		http.Error(w, `{"error":"failed to create project"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

//
// List Projects
//

// ListProjects godoc
// @Summary      List projects
// @Description  Retrieves all projects owned by the authenticated user
// @Tags         Projects
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   Project
// @Failure      500  {object}  map[string]string
// @Router       /projects [get]
func (h *Handler) ListProjects(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.Context().Value("userId").(string))
	if err != nil {
		http.Error(w, `{"error":"invalid user context"}`, http.StatusUnauthorized)
		return
	}

	projects, err := h.Service.ListProjects(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error":"failed to retrieve projects"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(projects)
}
