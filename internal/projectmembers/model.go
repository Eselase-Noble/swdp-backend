package projectmembers

import (
	"time"

	"github.com/google/uuid"
)

type ProjectMember struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProjectID uuid.UUID `gorm:"type:uuid;primaryKey"`
	Role      string    `gorm:"not null"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedYN bool `gorm:"default:false"`
}

func (ProjectMember) TableName() string {
	return "project_members"
}
