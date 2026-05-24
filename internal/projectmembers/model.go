package projectmembers

import (
	"time"

	"github.com/google/uuid"
)

type ProjectMember struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey"             json:"user_id"`
	ProjectID uuid.UUID `gorm:"type:uuid;primaryKey"             json:"project_id"`
	Role      string    `gorm:"not null"                         json:"role"`

	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedYN bool      `gorm:"column:deleted_yn;default:false"  json:"-"`
}

func (ProjectMember) TableName() string {
	return "project_members"
}
