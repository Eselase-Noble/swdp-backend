package projects

import (
	"context"

	"github.com/google/uuid"
)

//type Service interface {
//	CreateProject(ctx context.Context, ownerID uuid.UUID, name string) error
//	ListProjects(ctx context.Context, userID uuid.UUID) ([]Project, error)
//}

type Service struct {
	Repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) CreateProject(ctx context.Context, ownerID uuid.UUID, name string) error {
	project := &Project{
		Name:    name,
		OwnerID: ownerID,
		Version: 1,
	}
	return s.Repo.Create(ctx, project)
}

func (s *Service) ListProjects(ctx context.Context, userID uuid.UUID) ([]Project, error) {
	return s.Repo.FindByUser(ctx, userID)
}
