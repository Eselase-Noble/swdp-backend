package workspaces

import (
	"github.com/google/uuid"
)

type Service struct {
	Repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) CreateWorkspace(
	projectID, userID uuid.UUID,
) error {
	ws := &Workspace{
		ProjectID: projectID,
		UserID:    userID,
		Status:    WorkspaceCreated,
	}
	return s.Repo.Create(ws)
}

func (s *Service) StartWorkspace(id, userID uuid.UUID) error {
	ws, err := s.Repo.FindOwned(id, userID)
	if err != nil {
		return err
	}

	if ws.Status == WorkspaceActive {
		return nil
	}

	return s.Repo.UpdateStatus(ws.ID, WorkspaceActive)
}

func (s *Service) StopWorkspace(id, userID uuid.UUID) error {
	_, err := s.Repo.FindOwned(id, userID)
	if err != nil {
		return err
	}

	return s.Repo.UpdateStatus(id, WorkspaceStopped)
}

func (s *Service) DeleteWorkspace(id, userID uuid.UUID) error {
	_, err := s.Repo.FindOwned(id, userID)
	if err != nil {
		return err
	}

	return s.Repo.Delete(id)
}

func (s *Service) GetStatus(id, userID uuid.UUID) (WorkspaceStatus, error) {
	ws, err := s.Repo.FindOwned(id, userID)
	if err != nil {
		return "", err
	}
	return ws.Status, nil
}
