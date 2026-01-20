package users

import "github.com/google/uuid"

// Service provides user-related business logic.
type Service struct {
	Repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{Repo: repo}
}

// AddUser creates a user.
func (s *Service) AddUser(user *Users) error {
	return s.Repo.AddUser(user)
}

// GetUserByID retrieves a user by ID.
func (s *Service) GetUserByID(userID uuid.UUID) (*Users, error) {
	return s.Repo.GetUserByID(userID)
}

// GetAllUsers retrieves all users.
func (s *Service) GetAllUsers() ([]Users, error) {
	return s.Repo.GetAllUsers()
}

// UpdateUser updates a user.
func (s *Service) UpdateUser(userID uuid.UUID, updatedData map[string]interface{}) error {
	return s.Repo.UpdateUser(userID, updatedData)
}

// DeleteUser deletes a user.
func (s *Service) DeleteUser(userID uuid.UUID) error {
	return s.Repo.DeleteUser(userID)
}
