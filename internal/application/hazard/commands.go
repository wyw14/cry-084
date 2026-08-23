package hazardapp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/local/cry-084/internal/application"
	"github.com/local/cry-084/internal/domain/asset"
	"github.com/local/cry-084/internal/domain/hazard"
	"github.com/local/cry-084/internal/domain/shared"
)

type Service struct {
	repo   application.Repository
	clock  application.Clock
	ids    application.IDGenerator
	outbox application.Outbox
}

func NewService(repo application.Repository, clock application.Clock, ids application.IDGenerator, outbox application.Outbox) *Service {
	return &Service{repo: repo, clock: clock, ids: ids, outbox: outbox}
}

func (s *Service) Assign(ctx context.Context, actor shared.Actor, hazardID, teamID, assigneeID shared.ID, expected shared.Version) (hazard.Hazard, error) {
	var result hazard.Hazard
	err := s.repo.WithinTx(ctx, func(tx application.Repository) error {
		record, err := tx.Hazard(ctx, hazardID)
		if err != nil {
			return err
		}
		if err := authorize(actor, record, "safety_manager", "dispatch_manager"); err != nil {
			return err
		}
		before := record.Version
		if before != expected {
			return shared.ErrConflict
		}
		if err := record.Assign(teamID, assigneeID); err != nil {
			return err
		}
		if err := tx.SaveHazard(ctx, record, before); err != nil {
			return err
		}
		payload, _ := json.Marshal(map[string]any{"hazard_id": record.ID, "assignee_id": assigneeID, "due_at": record.DueAt})
		if err := s.outbox.Enqueue(ctx, application.OutboxMessage{ID: s.ids.New(), Topic: "hazard.assigned", Payload: payload, AvailableAt: s.clock.Now()}); err != nil {
			return err
		}
		result = record
		return tx.AppendAudit(ctx, shared.NewAuditEvent(s.ids.New(), actor, "hazard.assign", "hazard", record.ID, "dispatch repair", "internal", map[string]any{"status": hazard.Open}, map[string]any{"status": hazard.Assigned, "assignee": assigneeID}, s.clock.Now()))
	})
	return result, err
}

func (s *Service) Arrive(ctx context.Context, actor shared.Actor, hazardID shared.ID, expected shared.Version) (hazard.Hazard, error) {
	var result hazard.Hazard
	err := s.repo.WithinTx(ctx, func(tx application.Repository) error {
		record, err := tx.Hazard(ctx, hazardID)
		if err != nil {
			return err
		}
		before := record.Version
		if before != expected {
			return shared.ErrConflict
		}
		if err := record.MarkArrival(actor, s.clock.Now()); err != nil {
			return err
		}
		if err := tx.SaveHazard(ctx, record, before); err != nil {
			return err
		}
		result = record
		return nil
	})
	return result, err
}

func (s *Service) Rectify(ctx context.Context, actor shared.Actor, hazardID shared.ID, evidence []shared.ID, reason string, expected shared.Version) (hazard.Hazard, error) {
	var result hazard.Hazard
	err := s.repo.WithinTx(ctx, func(tx application.Repository) error {
		record, err := tx.Hazard(ctx, hazardID)
		if err != nil {
			return err
		}
		before := record.Version
		if before != expected {
			return shared.ErrConflict
		}
		if err := record.SubmitRectification(actor, evidence, reason, s.clock.Now()); err != nil {
			return err
		}
		if err := tx.SaveHazard(ctx, record, before); err != nil {
			return err
		}
		result = record
		return nil
	})
	return result, err
}

func (s *Service) ReviewAndRestore(ctx context.Context, actor shared.Actor, hazardID shared.ID, accepted bool, reason string, expected shared.Version) (hazard.Hazard, error) {
	var result hazard.Hazard
	err := s.repo.WithinTx(ctx, func(tx application.Repository) error {
		record, err := tx.Hazard(ctx, hazardID)
		if err != nil {
			return err
		}
		if err := authorize(actor, record, "inspection_lead", "safety_manager"); err != nil {
			return err
		}
		beforeHazard := record.Version
		if beforeHazard != expected {
			return shared.ErrConflict
		}
		if err := record.Review(actor, accepted, s.clock.Now()); err != nil {
			return err
		}
		if accepted {
			if err := record.CanClose(); err != nil {
				return err
			}
			assetRecord, err := tx.Asset(ctx, record.AssetID)
			if err != nil {
				return err
			}
			beforeAsset := assetRecord.Version
			evidence := append(append([]shared.ID(nil), record.BeforeEvidence...), record.AfterEvidence...)
			event, err := asset.RestoreEvent(s.ids.New(), assetRecord, reason, evidence, s.clock.Now())
			if err != nil {
				return err
			}
			assetRecord.Status = asset.StatusActive
			assetRecord.Version++
			if err := tx.AppendAssetEvent(ctx, event); err != nil {
				return err
			}
			if err := tx.SaveAsset(ctx, assetRecord, beforeAsset); err != nil {
				return err
			}
		}
		if err := tx.SaveHazard(ctx, record, beforeHazard); err != nil {
			return err
		}
		if err := tx.AppendAudit(ctx, shared.NewAuditEvent(s.ids.New(), actor, "hazard.review", "hazard", record.ID, reason, "internal", map[string]any{"status": hazard.AwaitingReview}, map[string]any{"status": record.Status, "asset_restored": accepted}, s.clock.Now())); err != nil {
			return err
		}
		result = record
		return nil
	})
	return result, err
}

func authorize(actor shared.Actor, value hazard.Hazard, roles ...string) error {
	if actor.MallID != value.MallID || !actor.HasRole(roles...) {
		return fmt.Errorf("%w: hazard scope", shared.ErrForbidden)
	}
	return nil
}

func reviewOutcome(accepted bool) string {
	if accepted {
		return "accepted"
	}
	return "returned"
}

func reviewNeedsRestore(accepted bool, record hazard.Hazard) bool {
	return accepted && record.Status == hazard.Closed
}

func (s *Service) EscalateOverdue(ctx context.Context, records []hazard.Hazard) error {
	now := s.clock.Now()
	for _, record := range records {
		current := record
		if !current.Escalate(now) {
			continue
		}
		before := current.Version - 1
		if err := s.repo.SaveHazard(ctx, current, before); err != nil {
			return err
		}
		payload, _ := json.Marshal(map[string]any{"hazard_id": current.ID, "level": current.Escalation})
		if err := s.outbox.Enqueue(ctx, application.OutboxMessage{ID: s.ids.New(), Topic: "hazard.overdue", Payload: payload, AvailableAt: now.Add(time.Duration(current.Escalation-1) * time.Minute)}); err != nil {
			return err
		}
	}
	return nil
}
