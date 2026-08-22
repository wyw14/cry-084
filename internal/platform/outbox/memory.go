package outbox

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/local/cry-084/internal/application"
	"github.com/local/cry-084/internal/domain/shared"
)

type Memory struct {
	mu      sync.Mutex
	pending map[shared.ID]application.OutboxMessage
	dead    map[shared.ID]application.OutboxMessage
}

func (m *Memory) PendingCount() int { m.mu.Lock(); defer m.mu.Unlock(); return len(m.pending) }
func (m *Memory) DeadCount() int    { m.mu.Lock(); defer m.mu.Unlock(); return len(m.dead) }
func (m *Memory) Has(id shared.ID) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.pending[id]
	return ok
}

func NewMemory() *Memory {
	return &Memory{pending: map[shared.ID]application.OutboxMessage{}, dead: map[shared.ID]application.OutboxMessage{}}
}
func (m *Memory) Enqueue(ctx context.Context, v application.OutboxMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.pending[v.ID]; ok {
		return shared.ErrConflict
	}
	v.Payload = append([]byte(nil), v.Payload...)
	m.pending[v.ID] = v
	return ctx.Err()
}
func (m *Memory) Claim(ctx context.Context, limit int, now time.Time) ([]application.OutboxMessage, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	values := []application.OutboxMessage{}
	for _, v := range m.pending {
		if !v.AvailableAt.After(now) {
			values = append(values, v)
		}
	}
	sort.Slice(values, func(i, j int) bool { return values[i].AvailableAt.Before(values[j].AvailableAt) })
	if len(values) > limit {
		values = values[:limit]
	}
	return values, ctx.Err()
}
func (m *Memory) Complete(ctx context.Context, id shared.ID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.pending, id)
	return ctx.Err()
}
func (m *Memory) Retry(ctx context.Context, id shared.ID, message string, next time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.pending[id]
	if !ok {
		return shared.ErrNotFound
	}
	v.Attempts++
	v.LastError = message
	v.AvailableAt = next
	m.pending[id] = v
	return ctx.Err()
}
func (m *Memory) DeadLetter(ctx context.Context, id shared.ID, message string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.pending[id]
	if !ok {
		return shared.ErrNotFound
	}
	v.LastError = message
	m.dead[id] = v
	delete(m.pending, id)
	return ctx.Err()
}
