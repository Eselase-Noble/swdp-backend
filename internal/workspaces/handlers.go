package workspaces

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"web-based-dev-platform-backend/internal/middleware"
	"web-based-dev-platform-backend/internal/runtime"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	Service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{Service: service}
}

func userIDFromCtx(r *http.Request) (uuid.UUID, error) {
	user := middleware.GetUser(r.Context())
	if user == nil {
		return uuid.Nil, http.ErrNoCookie
	}
	return uuid.Parse(user.Subject)
}

// fileRuntime returns the FileRuntime if the backing runtime supports it.
func (h *Handler) fileRuntime() (runtime.FileRuntime, bool) {
	fr, ok := h.Service.Runtime.(runtime.FileRuntime)
	return fr, ok
}

//
// List Workspaces
//

func (h *Handler) ListWorkspaces(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(r.URL.Query().Get("project_id"))
	if err != nil {
		http.Error(w, "invalid project_id", http.StatusBadRequest)
		return
	}

	userID, err := userIDFromCtx(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	wsList, err := h.Service.ListWorkspaces(r.Context(), projectID, userID)
	if err != nil {
		http.Error(w, "failed to list workspaces", http.StatusInternalServerError)
		return
	}

	// Return [] not null when empty
	if wsList == nil {
		wsList = []Workspace{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wsList)
}

//
// Create Workspace
//

// CreateWorkspace godoc
// @Summary      Create workspace
// @Description  Creates a workspace under a project for the authenticated user
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
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	userID, err := userIDFromCtx(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	ws, err := h.Service.CreateWorkspace(r.Context(), projectID, userID)
	if err != nil {
		http.Error(w, "failed to create workspace", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ws)
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
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid workspace id", http.StatusBadRequest)
		return
	}
	userID, err := userIDFromCtx(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.Service.StartWorkspace(r.Context(), id, userID); err != nil {
		if errors.Is(err, ErrNotOwned) {
			http.Error(w, "workspace not found", http.StatusForbidden)
		} else {
			http.Error(w, "workspace container is still provisioning — try again in a moment", http.StatusServiceUnavailable)
		}
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
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid workspace id", http.StatusBadRequest)
		return
	}
	userID, err := userIDFromCtx(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.Service.StopWorkspace(r.Context(), id, userID); err != nil {
		if errors.Is(err, ErrNotOwned) {
			http.Error(w, "workspace not found", http.StatusForbidden)
		} else {
			http.Error(w, "failed to stop workspace", http.StatusInternalServerError)
		}
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
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid workspace id", http.StatusBadRequest)
		return
	}
	userID, err := userIDFromCtx(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	status, err := h.Service.GetStatus(r.Context(), id, userID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": string(status)})
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
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid workspace id", http.StatusBadRequest)
		return
	}
	userID, err := userIDFromCtx(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.Service.DeleteWorkspace(r.Context(), id, userID); err != nil {
		if errors.Is(err, ErrNotOwned) {
			http.Error(w, "workspace not found", http.StatusForbidden)
		} else {
			http.Error(w, "failed to delete workspace", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ─── File API helpers ─────────────────────────────────────────────────────────

// workspaceIDAndOwner parses and validates the workspace ID and ownership.
func (h *Handler) workspaceIDAndOwner(r *http.Request) (uuid.UUID, uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return uuid.Nil, uuid.Nil, errors.New("invalid workspace id")
	}
	userID, err := userIDFromCtx(r)
	if err != nil {
		return uuid.Nil, uuid.Nil, errors.New("unauthorized")
	}
	if _, err := h.Service.Repo.FindOwned(id, userID); err != nil {
		return uuid.Nil, uuid.Nil, ErrNotOwned
	}
	return id, userID, nil
}

// filePath extracts and URL-decodes the wildcard filepath parameter.
func filePath(r *http.Request) string {
	p, _ := url.PathUnescape(chi.URLParam(r, "*"))
	return p
}

// ─── File Handlers ────────────────────────────────────────────────────────────

// ListFiles godoc
// @Summary      List workspace files
// @Tags         Workspaces
// @Security     BearerAuth
// @Param        id  path  string  true  "Workspace ID"
// @Success      200  {array}  runtime.FileEntry
// @Router       /workspaces/{id}/files [get]
func (h *Handler) ListFiles(w http.ResponseWriter, r *http.Request) {
	id, _, err := h.workspaceIDAndOwner(r)
	if err != nil {
		if errors.Is(err, ErrNotOwned) {
			http.Error(w, "workspace not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	fr, ok := h.fileRuntime()
	if !ok {
		http.Error(w, "file operations not supported by this runtime", http.StatusNotImplemented)
		return
	}

	entries, err := fr.ListFiles(r.Context(), id.String())
	if err != nil {
		http.Error(w, "failed to list files", http.StatusInternalServerError)
		return
	}
	if entries == nil {
		entries = []runtime.FileEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

// ReadFile godoc
// @Summary      Read a workspace file
// @Tags         Workspaces
// @Security     BearerAuth
// @Param        id        path  string  true  "Workspace ID"
// @Param        filepath  path  string  true  "File path relative to /workspace"
// @Success      200
// @Router       /workspaces/{id}/files/{filepath} [get]
func (h *Handler) ReadFile(w http.ResponseWriter, r *http.Request) {
	id, _, err := h.workspaceIDAndOwner(r)
	if err != nil {
		if errors.Is(err, ErrNotOwned) {
			http.Error(w, "workspace not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	fr, ok := h.fileRuntime()
	if !ok {
		http.Error(w, "file operations not supported by this runtime", http.StatusNotImplemented)
		return
	}

	content, err := fr.ReadFile(r.Context(), id.String(), filePath(r))
	if err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(content)
}

// WriteFile godoc
// @Summary      Create or overwrite a workspace file
// @Tags         Workspaces
// @Security     BearerAuth
// @Param        id        path  string  true  "Workspace ID"
// @Param        filepath  path  string  true  "File path relative to /workspace"
// @Success      204
// @Router       /workspaces/{id}/files/{filepath} [put]
func (h *Handler) WriteFile(w http.ResponseWriter, r *http.Request) {
	id, _, err := h.workspaceIDAndOwner(r)
	if err != nil {
		if errors.Is(err, ErrNotOwned) {
			http.Error(w, "workspace not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	fr, ok := h.fileRuntime()
	if !ok {
		http.Error(w, "file operations not supported by this runtime", http.StatusNotImplemented)
		return
	}

	content, err := io.ReadAll(io.LimitReader(r.Body, 10<<20)) // 10 MB cap
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	if err := fr.WriteFile(r.Context(), id.String(), filePath(r), content); err != nil {
		http.Error(w, "failed to write file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteFile godoc
// @Summary      Delete a workspace file or directory
// @Tags         Workspaces
// @Security     BearerAuth
// @Param        id        path  string  true  "Workspace ID"
// @Param        filepath  path  string  true  "File path relative to /workspace"
// @Success      204
// @Router       /workspaces/{id}/files/{filepath} [delete]
func (h *Handler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	id, _, err := h.workspaceIDAndOwner(r)
	if err != nil {
		if errors.Is(err, ErrNotOwned) {
			http.Error(w, "workspace not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	fr, ok := h.fileRuntime()
	if !ok {
		http.Error(w, "file operations not supported by this runtime", http.StatusNotImplemented)
		return
	}

	if err := fr.DeletePath(r.Context(), id.String(), filePath(r)); err != nil {
		http.Error(w, "failed to delete: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CreateDir godoc
// @Summary      Create a directory in the workspace
// @Tags         Workspaces
// @Security     BearerAuth
// @Param        id        path  string  true  "Workspace ID"
// @Param        filepath  path  string  true  "Directory path relative to /workspace"
// @Success      204
// @Router       /workspaces/{id}/dirs/{filepath} [post]
func (h *Handler) CreateDir(w http.ResponseWriter, r *http.Request) {
	id, _, err := h.workspaceIDAndOwner(r)
	if err != nil {
		if errors.Is(err, ErrNotOwned) {
			http.Error(w, "workspace not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	fr, ok := h.fileRuntime()
	if !ok {
		http.Error(w, "file operations not supported by this runtime", http.StatusNotImplemented)
		return
	}

	if err := fr.CreateDir(r.Context(), id.String(), filePath(r)); err != nil {
		http.Error(w, "failed to create directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
