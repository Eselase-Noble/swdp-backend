package workspaces

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

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
	err := r.DB.First(&ws, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &ws, nil
}

func (r *Repository) FindOwned(id, userID uuid.UUID) (*Workspace, error) {
	var ws Workspace
	err := r.DB.First(&ws, "id = ? AND user_id = ?", id, userID).Error
	if err != nil {
		return nil, errors.New("workspace not owned by user")
	}
	return &ws, nil
}

func (r *Repository) UpdateStatus(id uuid.UUID, status WorkspaceStatus) error {
	return r.DB.Model(&Workspace{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *Repository) Delete(id uuid.UUID) error {
	return r.DB.Delete(&Workspace{}, "id = ?", id).Error
}
