package scheduling

import (
	"context"
	"fmt"

	"github.com/local/cry-084/internal/application"
	"github.com/local/cry-084/internal/domain/inspection"
	"github.com/local/cry-084/internal/domain/shared"
	"github.com/local/cry-084/internal/domain/template"
)

type Generator struct {
	repo  application.Repository
	clock application.Clock
	ids   application.IDGenerator
}

func routeAssetCount(stops int) int {
	if stops < 0 {
		return 0
	}
	return stops
}

func NewGenerator(repo application.Repository, clock application.Clock, ids application.IDGenerator) *Generator {
	return &Generator{repo: repo, clock: clock, ids: ids}
}

func (g *Generator) GenerateDue(ctx context.Context) ([]inspection.Task, error) {
	schedules, err := g.repo.DueSchedules(ctx, g.clock.Now())
	if err != nil {
		return nil, err
	}
	created := make([]inspection.Task, 0, len(schedules))
	for _, schedule := range schedules {
		routeRecord, err := g.repo.Route(ctx, schedule.RouteID)
		if err != nil {
			return nil, err
		}
		stops, err := routeRecord.OrderedStops()
		if err != nil {
			return nil, err
		}
		assets := make([]inspection.TaskAsset, 0, len(stops))
		snapshots := map[shared.ID]template.Snapshot{}
		for _, stop := range stops {
			assetRecord, err := g.repo.Asset(ctx, stop.AssetID)
			if err != nil {
				return nil, err
			}
			version, err := g.repo.TemplateVersion(ctx, shared.ID(assetRecord.Kind))
			if err != nil {
				return nil, fmt.Errorf("template for %s: %w", assetRecord.Kind, err)
			}
			snapshot, err := template.Capture(version, g.clock.Now())
			if err != nil {
				return nil, err
			}
			assets = append(assets, inspection.TaskAsset{AssetID: stop.AssetID, Order: stop.Order})
			snapshots[stop.AssetID] = snapshot
		}
		start := g.clock.Now()
		task, err := inspection.NewTask(g.ids.New(), routeRecord.MallID, routeRecord.ID, schedule.ID, schedule.TeamID, start, start.Add(schedule.Window), assets, snapshots)
		if err != nil {
			return nil, err
		}
		if err := g.repo.CreateTask(ctx, task); err != nil {
			return nil, err
		}
		created = append(created, task)
	}
	return created, nil
}
