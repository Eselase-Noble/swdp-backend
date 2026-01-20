package workspaces

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	Service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{Service: service}
}

//
// Create Workspace
//

// CreateWorkspace godoc
// @Summary      Create workspace
// @Description  Creates a workspace under a project
// @Tags         Workspaces
// @Produce      json
// @Security     BearerAuth
// @Param        project_id  query  string  true  "Project ID"
// @Success      201
// @Failure      400  {object}  map[string]string
// @Router       /workspaces [post]
func (h *Handler) CreateWorkspace(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(r.URL.Query().Get("project_id"))
	if err != nil {
		http.Error(w, "invalid project id", 400)
		return
	}

	userID, _ := uuid.Parse(r.Context().Value("userId").(string))

	if err := h.Service.CreateWorkspace(r.Context(), projectID, userID); err != nil {
		http.Error(w, "failed to create workspace", 500)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

//
// Start Workspace
//

// StartWorkspace godoc
// @Summary      Start workspace
// @Tags         Workspaces
// @Security     BearerAuth
// @Param        id  path  string  true  "Workspace ID"
// @Success      200
// @Router       /workspaces/{id}/start [post]
func (h *Handler) StartWorkspace(w http.ResponseWriter, r *http.Request) {
	id, _ := uuid.Parse(chi.URLParam(r, "id"))
	userID, _ := uuid.Parse(r.Context().Value("userId").(string))

	if err := h.Service.StartWorkspace(id, userID); err != nil {
		http.Error(w, "forbidden", 403)
		return
	}

	w.WriteHeader(http.StatusOK)
}

//
// Stop Workspace
//

// StopWorkspace godoc
// @Summary      Stop workspace
// @Tags         Workspaces
// @Security     BearerAuth
// @Param        id  path  string  true  "Workspace ID"
// @Success      200
// @Router       /workspaces/{id}/stop [post]
func (h *Handler) StopWorkspace(w http.ResponseWriter, r *http.Request) {
	id, _ := uuid.Parse(chi.URLParam(r, "id"))
	userID, _ := uuid.Parse(r.Context().Value("userId").(string))

	if err := h.Service.StopWorkspace(id, userID); err != nil {
		http.Error(w, "forbidden", 403)
		return
	}

	w.WriteHeader(http.StatusOK)
}

//
// Workspace Status
//

// WorkspaceStatus godoc
// @Summary      Workspace status
// @Tags         Workspaces
// @Security     BearerAuth
// @Param        id  path  string  true  "Workspace ID"
// @Success      200  {object}  map[string]string
// @Router       /workspaces/{id}/status [get]
func (h *Handler) WorkspaceStatus(w http.ResponseWriter, r *http.Request) {
	id, _ := uuid.Parse(chi.URLParam(r, "id"))
	userID, _ := uuid.Parse(r.Context().Value("userId").(string))

	status, err := h.Service.GetStatus(id, userID)
	if err != nil {
		http.Error(w, "not found", 404)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status": string(status),
	})
}

//
// Delete Workspace
//

// DeleteWorkspace godoc
// @Summary      Delete workspace
// @Tags         Workspaces
// @Security     BearerAuth
// @Param        id  path  string  true  "Workspace ID"
// @Success      204
// @Router       /workspaces/{id} [delete]
func (h *Handler) DeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	id, _ := uuid.Parse(chi.URLParam(r, "id"))
	userID, _ := uuid.Parse(r.Context().Value("userId").(string))

	if err := h.Service.DeleteWorkspace(id, userID); err != nil {
		http.Error(w, "forbidden", 403)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
