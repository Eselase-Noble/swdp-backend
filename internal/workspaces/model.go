package workspaces

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WorkspaceStatus string

const (
	WorkspaceCreated WorkspaceStatus = "created"
	WorkspaceActive  WorkspaceStatus = "active"
	WorkspaceStopped WorkspaceStatus = "stopped"
)

type Workspace struct {
	ID        uuid.UUID       `gorm:"type:uuid;primaryKey"              json:"id"`
	ProjectID uuid.UUID       `gorm:"type:uuid;not null"                json:"project_id"`
	UserID    uuid.UUID       `gorm:"type:uuid;not null"                json:"user_id"`
	Status    WorkspaceStatus `gorm:"type:varchar(20)"                  json:"status"`

	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"  json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"  json:"updated_at"`
	DeletedYN bool      `gorm:"column:deleted_yn;default:false"   json:"-"`
}

func (w *Workspace) BeforeCreate(tx *gorm.DB) error {
	w.ID = uuid.New()
	return nil
}
