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
	// Always attempt the Docker start — don't trust DB status alone.
	// ContainerStart is idempotent: if the container is already running the
	// runtime returns a "not modified" error which we treat as success.
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
	ws, err := s.Repo.FindOwned(id, userID)
	if err != nil {
		return "", err
	}

	// Ask the runtime for the authoritative state, not just the DB record.
	rtStatus, err := s.Runtime.Status(ctx, id.String())
	if err != nil {
		// Runtime doesn't know about it (container not yet created, daemon
		// restart, etc.) — return the DB record as the best available answer.
		return ws.Status, nil
	}

	var actual WorkspaceStatus
	switch rtStatus {
	case runtime.StatusRunning:
		actual = WorkspaceActive
	case runtime.StatusStopped:
		actual = WorkspaceStopped
	default:
		actual = WorkspaceCreated
	}

	// Keep the DB in sync with the real container state.  If the Docker daemon
	// was restarted (or the container exited on its own) the DB might still say
	// "active" — this corrects that so the next StartWorkspace call works.
	if actual != ws.Status {
		_ = s.Repo.UpdateStatus(id, actual)
	}

	return actual, nil
}
