package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/local/cry-084/internal/application"
	"github.com/local/cry-084/internal/domain/asset"
	"github.com/local/cry-084/internal/domain/hazard"
	"github.com/local/cry-084/internal/domain/inspection"
	"github.com/local/cry-084/internal/domain/inventory"
	"github.com/local/cry-084/internal/domain/route"
	"github.com/local/cry-084/internal/domain/shared"
	"github.com/local/cry-084/internal/domain/template"
)

type data struct {
	assets       map[shared.ID]asset.Asset
	events       map[shared.ID][]asset.Event
	routes       map[shared.ID]route.Route
	schedules    map[shared.ID]route.Schedule
	templates    map[shared.ID]template.Version
	tasks        map[shared.ID]inspection.Task
	results      map[shared.ID]inspection.Result
	resultKeys   map[string]shared.ID
	hazards      map[shared.ID]hazard.Hazard
	parts        map[shared.ID]inventory.Part
	consumptions map[shared.ID]inventory.Consumption
	audits       []shared.AuditEvent
}

type Store struct {
	mu   sync.RWMutex
	data data
	tx   bool
}

func New() *Store { return &Store{data: newData()} }

func newData() data {
	return data{assets: map[shared.ID]asset.Asset{}, events: map[shared.ID][]asset.Event{}, routes: map[shared.ID]route.Route{}, schedules: map[shared.ID]route.Schedule{}, templates: map[shared.ID]template.Version{}, tasks: map[shared.ID]inspection.Task{}, results: map[shared.ID]inspection.Result{}, resultKeys: map[string]shared.ID{}, hazards: map[shared.ID]hazard.Hazard{}, parts: map[shared.ID]inventory.Part{}, consumptions: map[shared.ID]inventory.Consumption{}}
}

func (s *Store) WithinTx(ctx context.Context, fn func(application.Repository) error) error {
	if s.tx {
		return fn(s)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	working := &Store{data: cloneData(s.data), tx: true}
	if err := fn(working); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	s.data = working.data
	return nil
}

func (s *Store) Asset(ctx context.Context, id shared.ID) (asset.Asset, error) {
	s.rlock()
	defer s.runlock()
	value, ok := s.data.assets[id]
	if !ok {
		return asset.Asset{}, shared.ErrNotFound
	}
	return value, ctx.Err()
}
func (s *Store) SaveAsset(ctx context.Context, value asset.Asset, expected shared.Version) error {
	s.lock()
	defer s.unlock()
	existing, ok := s.data.assets[value.ID]
	if ok && existing.Version != expected {
		return shared.ErrConflict
	}
	s.data.assets[value.ID] = value
	return ctx.Err()
}
func (s *Store) AssetEvents(ctx context.Context, id shared.ID) ([]asset.Event, error) {
	s.rlock()
	defer s.runlock()
	return append([]asset.Event(nil), s.data.events[id]...), ctx.Err()
}
func (s *Store) AppendAssetEvent(ctx context.Context, value asset.Event) error {
	s.lock()
	defer s.unlock()
	s.data.events[value.AssetID] = append(s.data.events[value.AssetID], value)
	return ctx.Err()
}
func (s *Store) Route(ctx context.Context, id shared.ID) (route.Route, error) {
	s.rlock()
	defer s.runlock()
	value, ok := s.data.routes[id]
	if !ok {
		return route.Route{}, shared.ErrNotFound
	}
	value.Stops = append([]route.Stop(nil), value.Stops...)
	return value, ctx.Err()
}
func (s *Store) Schedule(ctx context.Context, id shared.ID) (route.Schedule, error) {
	s.rlock()
	defer s.runlock()
	value, ok := s.data.schedules[id]
	if !ok {
		return route.Schedule{}, shared.ErrNotFound
	}
	return value, ctx.Err()
}
func (s *Store) DueSchedules(ctx context.Context, at time.Time) ([]route.Schedule, error) {
	s.rlock()
	defer s.runlock()
	result := []route.Schedule{}
	for _, value := range s.data.schedules {
		if value.Active {
			result = append(result, value)
		}
	}
	return result, ctx.Err()
}
func (s *Store) TemplateVersion(ctx context.Context, id shared.ID) (template.Version, error) {
	s.rlock()
	defer s.runlock()
	value, ok := s.data.templates[id]
	if !ok {
		return template.Version{}, shared.ErrNotFound
	}
	value.Fields = append([]template.Field(nil), value.Fields...)
	return value, ctx.Err()
}
func (s *Store) Task(ctx context.Context, id shared.ID) (inspection.Task, error) {
	s.rlock()
	defer s.runlock()
	value, ok := s.data.tasks[id]
	if !ok {
		return inspection.Task{}, shared.ErrNotFound
	}
	return cloneTask(value), ctx.Err()
}
func (s *Store) SaveTask(ctx context.Context, value inspection.Task, expected shared.Version) error {
	s.lock()
	defer s.unlock()
	existing, ok := s.data.tasks[value.ID]
	if ok && existing.Version != expected {
		return shared.ErrConflict
	}
	s.data.tasks[value.ID] = cloneTask(value)
	return ctx.Err()
}
func (s *Store) CreateTask(ctx context.Context, value inspection.Task) error {
	s.lock()
	defer s.unlock()
	if _, ok := s.data.tasks[value.ID]; ok {
		return shared.ErrConflict
	}
	s.data.tasks[value.ID] = cloneTask(value)
	return ctx.Err()
}
func (s *Store) ResultByOfflineKey(ctx context.Context, key string) (inspection.Result, error) {
	s.rlock()
	defer s.runlock()
	id, ok := s.data.resultKeys[key]
	if !ok {
		return inspection.Result{}, shared.ErrNotFound
	}
	return s.data.results[id], ctx.Err()
}
func (s *Store) CreateResult(ctx context.Context, value inspection.Result) error {
	s.lock()
	defer s.unlock()
	if _, ok := s.data.resultKeys[value.OfflineKey]; ok {
		return shared.ErrConflict
	}
	s.data.results[value.ID] = value
	s.data.resultKeys[value.OfflineKey] = value.ID
	return ctx.Err()
}
func (s *Store) Hazard(ctx context.Context, id shared.ID) (hazard.Hazard, error) {
	s.rlock()
	defer s.runlock()
	value, ok := s.data.hazards[id]
	if !ok {
		return hazard.Hazard{}, shared.ErrNotFound
	}
	return cloneHazard(value), ctx.Err()
}
func (s *Store) SaveHazard(ctx context.Context, value hazard.Hazard, expected shared.Version) error {
	s.lock()
	defer s.unlock()
	existing, ok := s.data.hazards[value.ID]
	if ok && existing.Version != expected {
		return shared.ErrConflict
	}
	s.data.hazards[value.ID] = cloneHazard(value)
	return ctx.Err()
}
func (s *Store) CreateHazard(ctx context.Context, value hazard.Hazard) error {
	s.lock()
	defer s.unlock()
	if _, ok := s.data.hazards[value.ID]; ok {
		return shared.ErrConflict
	}
	s.data.hazards[value.ID] = cloneHazard(value)
	return ctx.Err()
}
func (s *Store) Part(ctx context.Context, id shared.ID) (inventory.Part, error) {
	s.rlock()
	defer s.runlock()
	value, ok := s.data.parts[id]
	if !ok {
		return inventory.Part{}, shared.ErrNotFound
	}
	return value, ctx.Err()
}
func (s *Store) SavePart(ctx context.Context, value inventory.Part, expected shared.Version) error {
	s.lock()
	defer s.unlock()
	existing, ok := s.data.parts[value.ID]
	if ok && existing.Version != expected {
		return shared.ErrConflict
	}
	s.data.parts[value.ID] = value
	return ctx.Err()
}
func (s *Store) CreateConsumption(ctx context.Context, value inventory.Consumption) error {
	s.lock()
	defer s.unlock()
	if _, ok := s.data.consumptions[value.ID]; ok {
		return shared.ErrConflict
	}
	s.data.consumptions[value.ID] = value
	return ctx.Err()
}
func (s *Store) AppendAudit(ctx context.Context, value shared.AuditEvent) error {
	s.lock()
	defer s.unlock()
	s.data.audits = append(s.data.audits, value)
	return ctx.Err()
}

func (s *Store) SeedAsset(value asset.Asset) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.assets[value.ID] = value
}
func (s *Store) SeedRoute(value route.Route) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.routes[value.ID] = value
}
func (s *Store) SeedSchedule(value route.Schedule) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.schedules[value.ID] = value
}
func (s *Store) SeedTemplate(key shared.ID, value template.Version) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.templates[key] = value
}
func (s *Store) SeedTask(value inspection.Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.tasks[value.ID] = cloneTask(value)
}
func (s *Store) SeedHazard(value hazard.Hazard) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.hazards[value.ID] = cloneHazard(value)
}
func (s *Store) SeedPart(value inventory.Part) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.parts[value.ID] = value
}
func (s *Store) Audits() []shared.AuditEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]shared.AuditEvent(nil), s.data.audits...)
}

func (s *Store) lock() {
	if !s.tx {
		s.mu.Lock()
	}
}
func (s *Store) unlock() {
	if !s.tx {
		s.mu.Unlock()
	}
}
func (s *Store) rlock() {
	if !s.tx {
		s.mu.RLock()
	}
}
func (s *Store) runlock() {
	if !s.tx {
		s.mu.RUnlock()
	}
}

func cloneTask(value inspection.Task) inspection.Task {
	value.Assets = append([]inspection.TaskAsset(nil), value.Assets...)
	snapshots := map[shared.ID]template.Snapshot{}
	for id, snapshot := range value.TemplateSnapshots {
		snapshot.Fields = append([]template.Field(nil), snapshot.Fields...)
		snapshots[id] = snapshot
	}
	value.TemplateSnapshots = snapshots
	return value
}
func cloneHazard(value hazard.Hazard) hazard.Hazard {
	value.BeforeEvidence = append([]shared.ID(nil), value.BeforeEvidence...)
	value.AfterEvidence = append([]shared.ID(nil), value.AfterEvidence...)
	return value
}
func cloneData(source data) data {
	target := newData()
	for id, v := range source.assets {
		target.assets[id] = v
	}
	for id, v := range source.events {
		target.events[id] = append([]asset.Event(nil), v...)
	}
	for id, v := range source.routes {
		v.Stops = append([]route.Stop(nil), v.Stops...)
		target.routes[id] = v
	}
	for id, v := range source.schedules {
		target.schedules[id] = v
	}
	for id, v := range source.templates {
		v.Fields = append([]template.Field(nil), v.Fields...)
		target.templates[id] = v
	}
	for id, v := range source.tasks {
		target.tasks[id] = cloneTask(v)
	}
	for id, v := range source.results {
		target.results[id] = v
	}
	for k, id := range source.resultKeys {
		target.resultKeys[k] = id
	}
	for id, v := range source.hazards {
		target.hazards[id] = cloneHazard(v)
	}
	for id, v := range source.parts {
		target.parts[id] = v
	}
	for id, v := range source.consumptions {
		target.consumptions[id] = v
	}
	target.audits = append([]shared.AuditEvent(nil), source.audits...)
	return target
}

func CheckVersion(actual, expected shared.Version) error {
	if actual != expected {
		return fmt.Errorf("%w: expected %d got %d", shared.ErrConflict, expected, actual)
	}
	return nil
}
