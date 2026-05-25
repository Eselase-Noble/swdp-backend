package workspaces

import (
	"context"
	"log"
	"web-based-dev-platform-backend/internal/runtime"

	"github.com/google/uuid"
)

type Service struct {
	Repo    *Repository
	Runtime runtime.Runtime
}

func NewService(repo *Repository, rt runtime.Runtime) *Service {
	return &Service{Repo: repo, Runtime: rt}
}

func (s *Service) ListWorkspaces(ctx context.Context, projectID, userID uuid.UUID) ([]Workspace, error) {
	return s.Repo.ListByProject(projectID, userID)
}

func (s *Service) CreateWorkspace(ctx context.Context, projectID, userID uuid.UUID) (*Workspace, error) {
	ws := &Workspace{
		ProjectID: projectID,
		UserID:    userID,
		Status:    WorkspaceCreated,
	}
	if err := s.Repo.Create(ws); err != nil {
		return nil, err
	}

	// Provision the Docker container in the background so the HTTP response
	// is not blocked by image pulls (ubuntu:22.04 is ~30 MB compressed).
	wsID := ws.ID
	go func() {
		if err := s.Runtime.Create(context.Background(), runtime.WorkspaceConfig{
			ID:       wsID.String(),
			Image:    "swdp-workspace:latest",
			CPUs:     1,
			MemoryMB: 512,
		}); err != nil {
			log.Printf("workspace runtime.Create %s: %v", wsID, err)
		}
	}()

	return ws, nil
}

func (s *Service) StartWorkspace(ctx context.Context, id, userID uuid.UUID) error {
	ws, err := s.Repo.FindOwned(id, userID)
	if err != nil {
		return err
	}
	if ws.Status == WorkspaceActive {
		return nil
	}
	if err := s.Runtime.Start(ctx, id.String()); err != nil {
		return err
	}
	return s.Repo.UpdateStatus(ws.ID, WorkspaceActive)
}

func (s *Service) StopWorkspace(ctx context.Context, id, userID uuid.UUID) error {
	_, err := s.Repo.FindOwned(id, userID)
	if err != nil {
		return err
	}
	if err := s.Runtime.Stop(ctx, id.String()); err != nil {
		return err
	}
	return s.Repo.UpdateStatus(id, WorkspaceStopped)
}

func (s *Service) DeleteWorkspace(ctx context.Context, id, userID uuid.UUID) error {
	_, err := s.Repo.FindOwned(id, userID)
	if err != nil {
		return err
	}
	if err := s.Runtime.Delete(ctx, id.String()); err != nil {
		return err
	}
	return s.Repo.Delete(id)
}

func (s *Service) GetStatus(ctx context.Context, id, userID uuid.UUID) (WorkspaceStatus, error) {
	if _, err := s.Repo.FindOwned(id, userID); err != nil {
		return "", err
	}

	// Ask the runtime for the authoritative state, not just the DB record.
	rtStatus, err := s.Runtime.Status(ctx, id.String())
	if err != nil {
		// Runtime doesn't know about it yet — fall back to DB.
		ws, dbErr := s.Repo.FindByID(id)
		if dbErr != nil {
			return "", dbErr
		}
		return ws.Status, nil
	}

	switch rtStatus {
	case runtime.StatusRunning:
		return WorkspaceActive, nil
	case runtime.StatusStopped:
		return WorkspaceStopped, nil
	default:
		return WorkspaceCreated, nil
	}
}
