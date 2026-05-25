package projects

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

//type Repository interface {
//	Create(ctx context.Context, project *Project) error
//	FindByUser(ctx context.Context, userID uuid.UUID) ([]Project, error)
//}

type Repository struct {
	DB *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) Create(ctx context.Context, project *Project) error {
	return r.DB.WithContext(ctx).Create(project).Error
}

func (r *Repository) FindByUser(ctx context.Context, userID uuid.UUID) ([]Project, error) {
	var projects []Project
	err := r.DB.WithContext(ctx).
		Where(
			"deleted_yn = false AND (owner_id = ? OR project_id IN (SELECT project_id FROM project_members WHERE user_id = ? AND deleted_yn = false))",
			userID, userID,
		).
		Find(&projects).Error
	return projects, err
}
