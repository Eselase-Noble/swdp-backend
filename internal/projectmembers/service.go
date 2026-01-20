package projectmembers

import (
	"github.com/google/uuid"
)

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
	member := &ProjectMember{
		ProjectID: projectID,
		UserID:    userID,
		Role:      role,
	}

	return s.Repo.Add(member)
}
