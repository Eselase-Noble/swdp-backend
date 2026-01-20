package projectmembers

import (
	"gorm.io/gorm"
)

//type Repository interface {
//	Add(ctx context.Context, member *ProjectMember) error
//}

type Repository struct {
	DB *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) Add(member *ProjectMember) error {
	return r.DB.Create(member).Error
}
