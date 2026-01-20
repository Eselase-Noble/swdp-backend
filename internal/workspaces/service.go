package workspaces

import (
	"context"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateWorkspace(
	ctx context.Context,
	projectID, userID uuid.UUID,
) error {
	ws := &Workspace{
		ProjectID: projectID,
		UserID:    userID,
		Status:    WorkspaceCreated,
	}
	return s.repo.Create(ws)
}

func (s *Service) StartWorkspace(id, userID uuid.UUID) error {
	ws, err := s.repo.FindOwned(id, userID)
	if err != nil {
		return err
	}

	if ws.Status == WorkspaceActive {
		return nil
	}

	return s.repo.UpdateStatus(ws.ID, WorkspaceActive)
}

func (s *Service) StopWorkspace(id, userID uuid.UUID) error {
	_, err := s.repo.FindOwned(id, userID)
	if err != nil {
		return err
	}

	return s.repo.UpdateStatus(id, WorkspaceStopped)
}

func (s *Service) DeleteWorkspace(id, userID uuid.UUID) error {
	_, err := s.repo.FindOwned(id, userID)
	if err != nil {
		return err
	}

	return s.repo.Delete(id)
}

func (s *Service) GetStatus(id, userID uuid.UUID) (WorkspaceStatus, error) {
	ws, err := s.repo.FindOwned(id, userID)
	if err != nil {
		return "", err
	}
	return ws.Status, nil
}
