package projectmembers

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Add(ctx context.Context, member *ProjectMember) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db}
}

func (r *repository) Add(ctx context.Context, member *ProjectMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}
