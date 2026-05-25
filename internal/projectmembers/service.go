package projectmembers

import (
	"errors"

	"github.com/google/uuid"
)

var ErrAlreadyMember = errors.New("user is already a member of this project")

//type Service interface {
//	AddMember(ctx context.Context, projectID, userID uuid.UUID, role string) error
//}

type Service struct {
	Repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) AddMember(projectID, userID uuid.UUID, role string) error {
	exists, err := s.Repo.Exists(projectID, userID)
	if err != nil {
		return err
	}
	if exists {
		return ErrAlreadyMember
	}

	member := &ProjectMember{
		ProjectID: projectID,
		UserID:    userID,
		Role:      role,
	}
	return s.Repo.Add(member)
}
