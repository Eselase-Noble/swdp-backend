package projects

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID          uuid.UUID `gorm:"type:uuid;column:project_id;default:gen_random_uuid();primaryKey" json:"id"`
	Name        string    `gorm:"column:project_name;not null"                                     json:"name"`
	OwnerID     uuid.UUID `gorm:"type:uuid;column:owner_id;not null"                               json:"owner_id"`
	Version     int       `gorm:"column:version"                                                   json:"version"`
	Environment string    `gorm:"type:environment_enum;column:environment;default:'dev'"           json:"environment"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"                                 json:"updated_at"`
	DeletedYN   bool      `gorm:"column:deleted_yn;default:false"                                  json:"-"`
}

func (Project) TableName() string {
	return "projects"
}
