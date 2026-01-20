package projects

import (
	"time"
	"web-based-dev-platform-backend/internal/users"

	"github.com/google/uuid"
)

type Project struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name        string    `gorm:"not null" json:"name"`
	OwnerID     uuid.UUID `gorm:"type:uuid;not null" json:"owner_id"`
	Version     int       `json:"version"`
	Environment string    `gorm:"type:environment_enum;default:'dev'" json:"environment"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedYN users.DeleteYn `gorm:"default:N"`
}

func (Project) TableName() string {
	return "projects"
}
