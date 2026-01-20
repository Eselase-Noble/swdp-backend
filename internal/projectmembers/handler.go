package projectmembers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type Handler struct {
	Service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{Service: service}
}

type AddProjectMemberRequest struct {
	UserID uuid.UUID `json:"user_id"`
	Role   string    `json:"role"`
}

//
// Add Project Member
//

// AddProjectMember godoc
// @Summary      Add project member
// @Description  Adds a user to a project with a role
// @Tags         Project Members
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        project_id  path      string                   true  "Project ID"
// @Param        body        body      AddProjectMemberRequest  true  "Member payload"
// @Success      201
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /projects/{project_id}/members [post]
func (h *Handler) AddProjectMember(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(r.URL.Query().Get("project_id"))
	if err != nil {
		http.Error(w, `{"error":"invalid project id"}`, http.StatusBadRequest)
		return
	}

	var req AddProjectMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	if err := h.Service.AddMember(r.Context(), projectID, req.UserID, req.Role); err != nil {
		http.Error(w, `{"error":"failed to add project member"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
