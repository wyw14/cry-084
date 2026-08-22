package notification

import (
	"context"
	"sync"

	"github.com/local/cry-084/internal/application"
)

type Local struct {
	mu     sync.Mutex
	values []application.Notification
}

func (l *Local) Send(ctx context.Context, v application.Notification) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.values = append(l.values, v)
	return ctx.Err()
}
func (l *Local) Values() []application.Notification {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]application.Notification(nil), l.values...)
}
