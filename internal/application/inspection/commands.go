package inspectionapp

import (
	"context"
	"fmt"
	"time"

	"github.com/local/cry-084/internal/application"
	"github.com/local/cry-084/internal/domain/hazard"
	inspectiondomain "github.com/local/cry-084/internal/domain/inspection"
	"github.com/local/cry-084/internal/domain/shared"
)

type Service struct {
	repo  application.Repository
	clock application.Clock
	ids   application.IDGenerator
}

func NewService(repo application.Repository, clock application.Clock, ids application.IDGenerator) *Service {
	return &Service{repo: repo, clock: clock, ids: ids}
}

type ClaimCommand struct {
	TaskID   shared.ID
	Actor    shared.Actor
	Expected shared.Version
}

func (s *Service) Claim(ctx context.Context, cmd ClaimCommand) (inspectiondomain.Task, error) {
	var result inspectiondomain.Task
	err := s.repo.WithinTx(ctx, func(tx application.Repository) error {
		task, err := tx.Task(ctx, cmd.TaskID)
		if err != nil {
			return err
		}
		before := task.Version
		if before != cmd.Expected {
			return shared.ErrConflict
		}
		if err := task.Claim(cmd.Actor); err != nil {
			return err
		}
		if err := tx.SaveTask(ctx, task, before); err != nil {
			return err
		}
		if err := tx.AppendAudit(ctx, shared.NewAuditEvent(s.ids.New(), cmd.Actor, "task.claim", "inspection_task", task.ID, "claim route task", requestID(ctx), map[string]any{"status": "pending"}, map[string]any{"status": task.Status}, s.clock.Now())); err != nil {
			return err
		}
		result = task
		return nil
	})
	return result, err
}

type BeginCommand struct {
	TaskID   shared.ID
	Actor    shared.Actor
	Expected shared.Version
}

func (s *Service) Begin(ctx context.Context, cmd BeginCommand) (inspectiondomain.Task, error) {
	var result inspectiondomain.Task
	err := s.repo.WithinTx(ctx, func(tx application.Repository) error {
		task, err := tx.Task(ctx, cmd.TaskID)
		if err != nil {
			return err
		}
		before := task.Version
		if before != cmd.Expected {
			return shared.ErrConflict
		}
		if err := task.Begin(cmd.Actor, s.clock.Now()); err != nil {
			return err
		}
		if err := tx.SaveTask(ctx, task, before); err != nil {
			return err
		}
		result = task
		return nil
	})
	return result, err
}

type SubmitCommand struct {
	TaskID    shared.ID
	Actor     shared.Actor
	AssetID   shared.ID
	QRSecret  string
	ScannedAt time.Time
	Kind      inspectiondomain.ResultKind
	Answers   map[string]any
	Evidence  []inspectiondomain.Evidence
	Expected  shared.Version
}

func (s *Service) Submit(ctx context.Context, cmd SubmitCommand) (inspectiondomain.Result, error) {
	var result inspectiondomain.Result
	err := s.repo.WithinTx(ctx, func(tx application.Repository) error {
		task, err := tx.Task(ctx, cmd.TaskID)
		if err != nil {
			return err
		}
		if task.Version != cmd.Expected || task.InspectorID != cmd.Actor.ID {
			return shared.ErrConflict
		}
		assetRecord, err := tx.Asset(ctx, cmd.AssetID)
		if err != nil {
			return err
		}
		scan := inspectiondomain.Scan{TaskID: task.ID, AssetID: cmd.AssetID, InspectorID: cmd.Actor.ID, QRSecret: cmd.QRSecret, ScannedAt: cmd.ScannedAt}
		if err := inspectiondomain.VerifyScan(task, assetRecord.ID, assetRecord.QRSecret, scan, s.clock.Now()); err != nil {
			return err
		}
		offlineKey := string(task.ID) + ":" + string(cmd.AssetID)
		if existing, err := tx.ResultByOfflineKey(ctx, offlineKey); err == nil {
			result = existing
			return nil
		}
		created, err := inspectiondomain.NewResult(s.ids.New(), task, scan, cmd.Kind, cmd.Answers, cmd.Evidence, s.clock.Now())
		if err != nil {
			return err
		}
		if err := tx.CreateResult(ctx, created); err != nil {
			return err
		}
		before := task.Version
		if err := task.MarkChecked(cmd.AssetID); err != nil {
			return err
		}
		if err := tx.SaveTask(ctx, task, before); err != nil {
			return err
		}
		if cmd.Kind != inspectiondomain.ResultNormal {
			priority := hazard.PriorityHigh
			if cmd.Kind == inspectiondomain.ResultDamaged {
				priority = hazard.PriorityCritical
			}
			beforeEvidence := make([]shared.ID, 0, len(cmd.Evidence))
			for _, evidence := range cmd.Evidence {
				beforeEvidence = append(beforeEvidence, evidence.ID)
			}
			createdHazard, err := hazard.New(s.ids.New(), task.MallID, cmd.AssetID, created.ID, priority, s.clock.Now().Add(8*time.Hour), beforeEvidence)
			if err != nil {
				return err
			}
			if err := tx.CreateHazard(ctx, createdHazard); err != nil {
				return err
			}
		}
		result = created
		return tx.AppendAudit(ctx, shared.NewAuditEvent(s.ids.New(), cmd.Actor, "inspection.submit", "asset", cmd.AssetID, "scan result", requestID(ctx), nil, map[string]any{"result": cmd.Kind}, s.clock.Now()))
	})
	return result, err
}

func requestID(ctx context.Context) string {
	if value, ok := ctx.Value("request_id").(string); ok {
		return value
	}
	return "internal"
}

func ValidateTaskWindow(start, end time.Time) error {
	if !end.After(start) || end.Sub(start) > 24*time.Hour {
		return fmt.Errorf("%w: invalid task time window", shared.ErrValidation)
	}
	return nil
}
