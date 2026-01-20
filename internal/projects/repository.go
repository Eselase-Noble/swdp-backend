package projects

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, project *Project) error
	FindByUser(ctx context.Context, userID uuid.UUID) ([]Project, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db}
}

func (r *repository) Create(ctx context.Context, project *Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

func (r *repository) FindByUser(ctx context.Context, userID uuid.UUID) ([]Project, error) {
	var projects []Project
	err := r.db.WithContext(ctx).
		Where("owner_id = ? AND deleted_yn = false", userID).
		Find(&projects).Error
	return projects, err
}
