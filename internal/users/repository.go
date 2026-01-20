package users

import (
	"fmt"
	"strconv"

	"gorm.io/gorm"
)

// Repository handles database operations for users.
type Repository struct {
	DB *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{DB: db}
}

// AddUser creates a new user record.
func (r *Repository) AddUser(user *User) error {
	user.userId = generateUserId(r)
	user.deleteYn = DeleteNo
	return r.DB.Create(user).Error
}

// GetUserByID retrieves a user by ID if not deleted.
func (r *Repository) GetUserByID(userID string) (*User, error) {
	var user User
	err := r.DB.
		Where("userId = ? AND deleteYn = ?", userID, DeleteNo).
		First(&user).Error
	return &user, err
}

// GetAllUsers retrieves all non-deleted users.
func (r *Repository) GetAllUsers() ([]User, error) {
	var users []User
	err := r.DB.
		Where("delete_yn = ?", DeleteNo).
		Find(&users).Error
	return users, err
}

// UpdateUser updates allowed fields of a user.
func (r *Repository) UpdateUser(userID string, updatedData map[string]interface{}) error {
	return r.DB.
		Model(&User{}).
		Where("user_id = ? AND delete_yn = ?", userID, DeleteNo).
		Updates(updatedData).Error
}

// DeleteUser soft-deletes a user.
func (r *Repository) DeleteUser(userID string) error {
	return r.DB.
		Model(&User{}).
		Where("user_id = ?", userID).
		Update("delete_yn", DeleteYes).Error
}

func generateUserId(r *Repository) string {
	const prefix = "ACL"

	var latestUserId string

	err := r.DB.
		Model(&User{}).
		Select("user_id").
		Where("delete_yn = ?", DeleteNo).
		Order("user_id DESC").
		Limit(1).
		Scan(&latestUserId).Error

	// If no user exists yet
	if err != nil || latestUserId == "" {
		return prefix + "00001"
	}

	// Extract numeric part
	numPart := latestUserId[len(prefix):]
	num, err := strconv.Atoi(numPart)
	if err != nil {
		return prefix + "00001"
	}

	return fmt.Sprintf("%s%05d", prefix, num+1)
}
