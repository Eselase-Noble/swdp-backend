package users

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	DB *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) AddUser(user *User) error {
	user.UserID = uuid.New()
	user.DeletedYN = false
	return r.DB.Create(user).Error
}

func (r *Repository) GetUserByID(userID string) (*User, error) {
	var user User
	err := r.DB.
		Where("user_id = ? AND deleted_yn = ?", userID, false).
		First(&user).Error
	return &user, err
}

func (r *Repository) GetAllUsers() ([]User, error) {
	var users []User
	err := r.DB.Where("deleted_yn = ?", false).Find(&users).Error
	return users, err
}

func (r *Repository) UpdateUser(userID string, updatedData map[string]interface{}) error {
	return r.DB.
		Model(&User{}).
		Where("user_id = ? AND deleted_yn = ?", userID, false).
		Updates(updatedData).Error
}

func (r *Repository) DeleteUser(userID string) error {
	return r.DB.
		Model(&User{}).
		Where("user_id = ?", userID).
		Update("deleted_yn", true).Error
}
