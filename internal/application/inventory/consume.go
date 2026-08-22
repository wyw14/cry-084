package inventoryapp

import (
	"context"

	"github.com/local/cry-084/internal/application"
	"github.com/local/cry-084/internal/domain/hazard"
	"github.com/local/cry-084/internal/domain/inventory"
	"github.com/local/cry-084/internal/domain/shared"
)

type Service struct {
	repo  application.Repository
	ids   application.IDGenerator
	clock application.Clock
}

func NewService(repo application.Repository, ids application.IDGenerator, clock application.Clock) *Service {
	return &Service{repo: repo, ids: ids, clock: clock}
}

func (s *Service) Consume(ctx context.Context, actor shared.Actor, hazardID, partID shared.ID, quantity int, expectedPart shared.Version) (inventory.Consumption, error) {
	var result inventory.Consumption
	err := s.repo.WithinTx(ctx, func(tx application.Repository) error {
		hazardRecord, err := tx.Hazard(ctx, hazardID)
		if err != nil {
			return err
		}
		if actor.MallID != hazardRecord.MallID || actor.ID != hazardRecord.AssigneeID || (hazardRecord.Status != hazard.Arrived && hazardRecord.Status != hazard.Rectifying) {
			return shared.ErrForbidden
		}
		part, err := tx.Part(ctx, partID)
		if err != nil {
			return err
		}
		if part.MallID != actor.MallID || part.Version != expectedPart {
			return shared.ErrConflict
		}
		before := part.Version
		consumed, err := part.Consume(s.ids.New(), hazardID, actor.ID, quantity)
		if err != nil {
			return err
		}
		if err := tx.SavePart(ctx, part, before); err != nil {
			return err
		}
		if err := tx.CreateConsumption(ctx, consumed); err != nil {
			return err
		}
		result = consumed
		return tx.AppendAudit(ctx, shared.NewAuditEvent(s.ids.New(), actor, "part.consume", "hazard", hazardID, "repair consumption", "internal", nil, map[string]any{"part_id": partID, "quantity": quantity}, s.clock.Now()))
	})
	return result, err
}
