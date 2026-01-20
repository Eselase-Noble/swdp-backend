package users

// Service provides user-related business logic.
type Service struct {
	Repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{Repo: repo}
}

// AddUser creates a user.
func (s *Service) AddUser(user *User) error {
	return s.Repo.AddUser(user)
}

// GetUserByID retrieves a user by ID.
func (s *Service) GetUserByID(userID string) (*User, error) {
	return s.Repo.GetUserByID(userID)
}

// GetAllUsers retrieves all users.
func (s *Service) GetAllUsers() ([]User, error) {
	return s.Repo.GetAllUsers()
}

// UpdateUser updates a user.
func (s *Service) UpdateUser(userID string, updatedData map[string]interface{}) error {
	return s.Repo.UpdateUser(userID, updatedData)
}

// DeleteUser deletes a user.
func (s *Service) DeleteUser(userID string) error {
	return s.Repo.DeleteUser(userID)
}
