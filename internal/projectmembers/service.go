package projectmembers

import (
	"context"

	"github.com/google/uuid"
)

type Service interface {
	AddMember(ctx context.Context, projectID, userID uuid.UUID, role string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo}
}

func (s *service) AddMember(ctx context.Context, projectID, userID uuid.UUID, role string) error {
	return s.repo.Add(ctx, &ProjectMember{
		ProjectID: projectID,
		UserID:    userID,
		Role:      role,
	})
}
