package users

import "time"

type DeleteYn string

const (
	DeleteYes DeleteYn = "Y"
	DeleteNo  DeleteYn = "N"
)

type User struct {
	userId    string   `gorm:"primary_key"`
	username  string   `gorm:"unique;size:50"`
	email     string   `gorm:"size:100;unique;not null"`
	password  string   `gorm:"size:100"`
	deleteYn  DeleteYn `gorm:"type:char(1);default:N"`
	createdAt time.Time
	updatedAt time.Time
	createdBy string `gorm:"size:100"`
	updatedBy string `gorm:"size:100"`
}
