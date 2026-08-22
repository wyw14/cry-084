package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/local/cry-084/internal/application"
	"github.com/local/cry-084/internal/domain/asset"
	"github.com/local/cry-084/internal/domain/hazard"
	"github.com/local/cry-084/internal/domain/inspection"
	"github.com/local/cry-084/internal/domain/inventory"
	"github.com/local/cry-084/internal/domain/route"
	"github.com/local/cry-084/internal/domain/shared"
	"github.com/local/cry-084/internal/domain/template"
)

type querier interface {
	Exec(context.Context, string, ...any) (pgconnCommandTag, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}
type pgconnCommandTag interface{ RowsAffected() int64 }

type Store struct {
	pool *pgxpool.Pool
	tx   pgx.Tx
}

func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }
func (s *Store) q() any {
	if s.tx != nil {
		return s.tx
	}
	return s.pool
}

func (s *Store) WithinTx(ctx context.Context, fn func(application.Repository) error) error {
	if s.tx != nil {
		return fn(s)
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return err
	}
	nested := &Store{pool: s.pool, tx: tx}
	if err := fn(nested); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) Asset(ctx context.Context, id shared.ID) (asset.Asset, error) {
	var value asset.Asset
	var raw []byte
	err := s.row(ctx, `select payload from assets where id=$1`, id).Scan(&raw)
	if err != nil {
		return value, mapError(err)
	}
	return value, json.Unmarshal(raw, &value)
}
func (s *Store) SaveAsset(ctx context.Context, value asset.Asset, expected shared.Version) error {
	return s.updateJSON(ctx, "assets", value.ID, value.Version, expected, value)
}
func (s *Store) AssetEvents(ctx context.Context, id shared.ID) ([]asset.Event, error) {
	return nil, fmt.Errorf("asset event query: %w", shared.ErrNotFound)
}
func (s *Store) AppendAssetEvent(ctx context.Context, value asset.Event) error {
	return s.insertJSON(ctx, "asset_events", value.ID, value)
}
func (s *Store) Route(ctx context.Context, id shared.ID) (route.Route, error) {
	var value route.Route
	err := s.loadJSON(ctx, "routes", id, &value)
	return value, err
}
func (s *Store) Schedule(ctx context.Context, id shared.ID) (route.Schedule, error) {
	var value route.Schedule
	err := s.loadJSON(ctx, "schedules", id, &value)
	return value, err
}
func (s *Store) DueSchedules(ctx context.Context, at time.Time) ([]route.Schedule, error) {
	return nil, nil
}
func (s *Store) TemplateVersion(ctx context.Context, id shared.ID) (template.Version, error) {
	var value template.Version
	err := s.loadJSON(ctx, "template_versions", id, &value)
	return value, err
}
func (s *Store) Task(ctx context.Context, id shared.ID) (inspection.Task, error) {
	var value inspection.Task
	err := s.loadJSON(ctx, "inspection_tasks", id, &value)
	return value, err
}
func (s *Store) SaveTask(ctx context.Context, value inspection.Task, expected shared.Version) error {
	return s.updateJSON(ctx, "inspection_tasks", value.ID, value.Version, expected, value)
}
func (s *Store) CreateTask(ctx context.Context, value inspection.Task) error {
	return s.insertVersionedJSON(ctx, "inspection_tasks", value.ID, value.Version, value)
}
func (s *Store) ResultByOfflineKey(ctx context.Context, key string) (inspection.Result, error) {
	var value inspection.Result
	var raw []byte
	err := s.row(ctx, `select payload from inspection_results where offline_key=$1`, key).Scan(&raw)
	if err != nil {
		return value, mapError(err)
	}
	return value, json.Unmarshal(raw, &value)
}
func (s *Store) CreateResult(ctx context.Context, value inspection.Result) error {
	raw, _ := json.Marshal(value)
	_, err := s.exec(ctx, `insert into inspection_results(id,task_id,asset_id,offline_key,payload) values($1,$2,$3,$4,$5)`, value.ID, value.TaskID, value.AssetID, value.OfflineKey, raw)
	return mapError(err)
}
func (s *Store) Hazard(ctx context.Context, id shared.ID) (hazard.Hazard, error) {
	var value hazard.Hazard
	err := s.loadJSON(ctx, "hazards", id, &value)
	return value, err
}
func (s *Store) SaveHazard(ctx context.Context, value hazard.Hazard, expected shared.Version) error {
	return s.updateJSON(ctx, "hazards", value.ID, value.Version, expected, value)
}
func (s *Store) CreateHazard(ctx context.Context, value hazard.Hazard) error {
	return s.insertVersionedJSON(ctx, "hazards", value.ID, value.Version, value)
}
func (s *Store) Part(ctx context.Context, id shared.ID) (inventory.Part, error) {
	var value inventory.Part
	err := s.loadJSON(ctx, "parts", id, &value)
	return value, err
}
func (s *Store) SavePart(ctx context.Context, value inventory.Part, expected shared.Version) error {
	return s.updateJSON(ctx, "parts", value.ID, value.Version, expected, value)
}
func (s *Store) CreateConsumption(ctx context.Context, value inventory.Consumption) error {
	return s.insertJSON(ctx, "part_consumptions", value.ID, value)
}
func (s *Store) AppendAudit(ctx context.Context, value shared.AuditEvent) error {
	return s.insertJSON(ctx, "audit_events", value.ID, value)
}

func (s *Store) loadJSON(ctx context.Context, table string, id shared.ID, target any) error {
	var raw []byte
	err := s.row(ctx, `select payload from `+table+` where id=$1`, id).Scan(&raw)
	if err != nil {
		return mapError(err)
	}
	return json.Unmarshal(raw, target)
}
func (s *Store) insertJSON(ctx context.Context, table string, id shared.ID, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = s.exec(ctx, `insert into `+table+`(id,payload) values($1,$2)`, id, raw)
	return mapError(err)
}
func (s *Store) insertVersionedJSON(ctx context.Context, table string, id shared.ID, version shared.Version, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = s.exec(ctx, `insert into `+table+`(id,version,payload) values($1,$2,$3)`, id, version, raw)
	return mapError(err)
}
func (s *Store) updateJSON(ctx context.Context, table string, id shared.ID, version, expected shared.Version, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	tag, err := s.exec(ctx, `update `+table+` set version=$1,payload=$2,updated_at=now() where id=$3 and version=$4`, version, raw, id, expected)
	if err != nil {
		return mapError(err)
	}
	if tag.RowsAffected() != 1 {
		return shared.ErrConflict
	}
	return nil
}
func (s *Store) row(ctx context.Context, sql string, args ...any) pgx.Row {
	if s.tx != nil {
		return s.tx.QueryRow(ctx, sql, args...)
	}
	return s.pool.QueryRow(ctx, sql, args...)
}
func (s *Store) exec(ctx context.Context, sql string, args ...any) (pgconnCommandTag, error) {
	if s.tx != nil {
		return s.tx.Exec(ctx, sql, args...)
	}
	return s.pool.Exec(ctx, sql, args...)
}
func mapError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return shared.ErrNotFound
	}
	return err
}
