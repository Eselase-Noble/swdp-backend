package projects

import (
	"context"

	"github.com/google/uuid"
)

type Service interface {
	CreateProject(ctx context.Context, ownerID uuid.UUID, name string) error
	ListProjects(ctx context.Context, userID uuid.UUID) ([]Project, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo}
}

func (s *service) CreateProject(ctx context.Context, ownerID uuid.UUID, name string) error {
	project := &Project{
		Name:    name,
		OwnerID: ownerID,
		Version: 1,
	}
	return s.repo.Create(ctx, project)
}

func (s *service) ListProjects(ctx context.Context, userID uuid.UUID) ([]Project, error) {
	return s.repo.FindByUser(ctx, userID)
}
