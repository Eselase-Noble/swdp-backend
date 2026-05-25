package projectmembers

import (
	"github.com/google/uuid"
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

func (r *Repository) Exists(projectID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.DB.Model(&ProjectMember{}).
		Where("project_id = ? AND user_id = ? AND deleted_yn = false", projectID, userID).
		Count(&count).Error
	return count > 0, err
}
