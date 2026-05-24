package users

import (
	"time"

	"github.com/google/uuid"
)

// User maps exactly to the `users` table defined in migrations/001_init.sql.
type User struct {
	UserID       uuid.UUID  `gorm:"type:uuid;column:user_id;primaryKey;default:gen_random_uuid()" json:"user_id"`
	Username     string     `gorm:"column:username;uniqueIndex;not null"                           json:"username"`
	Name         string     `gorm:"column:name;not null"                                           json:"name"`
	Email        string     `gorm:"column:email;uniqueIndex;not null"                              json:"email"`
	PasswordHash string     `gorm:"column:password_hash;not null"                                  json:"-"`
	Role         string     `gorm:"column:role;not null"                                           json:"role"`
	CreatedBy    *uuid.UUID `gorm:"column:created_by;type:uuid"                                    json:"created_by,omitempty"`
	UpdatedBy    *uuid.UUID `gorm:"column:updated_by;type:uuid"                                    json:"updated_by,omitempty"`
	CreatedAt    time.Time  `gorm:"column:created_at;autoCreateTime"                               json:"created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;autoUpdateTime"                               json:"updated_at"`
	DeletedYN    bool       `gorm:"column:deleted_yn;default:false"                                json:"-"`
}

func (User) TableName() string { return "users" }
