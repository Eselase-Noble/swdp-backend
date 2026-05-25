package workspaces

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrNotOwned is returned when the workspace does not exist or belongs to a different user.
var ErrNotOwned = errors.New("workspace not found or not owned by user")

type Repository struct {
	DB *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) Create(ws *Workspace) error {
	return r.DB.Create(ws).Error
}

func (r *Repository) FindByID(id uuid.UUID) (*Workspace, error) {
	var ws Workspace
	err := r.DB.First(&ws, "id = ? AND deleted_yn = false", id).Error
	if err != nil {
		return nil, err
	}
	return &ws, nil
}

func (r *Repository) FindOwned(id, userID uuid.UUID) (*Workspace, error) {
	var ws Workspace
	err := r.DB.First(&ws, "id = ? AND user_id = ? AND deleted_yn = false", id, userID).Error
	if err != nil {
		return nil, ErrNotOwned
	}
	return &ws, nil
}

func (r *Repository) UpdateStatus(id uuid.UUID, status WorkspaceStatus) error {
	return r.DB.Model(&Workspace{}).
		Where("id = ? AND deleted_yn = false", id).
		Update("status", status).Error
}

func (r *Repository) Delete(id uuid.UUID) error {
	return r.DB.Model(&Workspace{}).
		Where("id = ?", id).
		Update("deleted_yn", true).Error
}

func (r *Repository) ListByProject(projectID, userID uuid.UUID) ([]Workspace, error) {
	var ws []Workspace
	err := r.DB.Where("project_id = ? AND user_id = ? AND deleted_yn = false", projectID, userID).
		Order("created_at DESC").
		Find(&ws).Error
	return ws, err
}
