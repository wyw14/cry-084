package outbox

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/local/cry-084/internal/application"
)

type Handler interface {
	Handle(context.Context, application.OutboxMessage) error
}
type Worker struct {
	queue       application.Outbox
	handler     Handler
	maxAttempts int
	interval    time.Duration
}

func NewWorker(queue application.Outbox, handler Handler, maxAttempts int, interval time.Duration) *Worker {
	return &Worker{queue: queue, handler: handler, maxAttempts: maxAttempts, interval: interval}
}
func (w *Worker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case now := <-ticker.C:
			if err := w.tick(ctx, now); err != nil {
				return err
			}
		}
	}
}
func (w *Worker) tick(ctx context.Context, now time.Time) error {
	messages, err := w.queue.Claim(ctx, 20, now)
	if err != nil {
		return err
	}
	for _, message := range messages {
		err = w.handler.Handle(ctx, message)
		if err == nil {
			if doneErr := w.queue.Complete(ctx, message.ID); doneErr != nil {
				return doneErr
			}
			continue
		}
		if message.Attempts+1 >= w.maxAttempts {
			if deadErr := w.queue.DeadLetter(ctx, message.ID, err.Error()); deadErr != nil {
				return deadErr
			}
			continue
		}
		delay := time.Duration(math.Pow(2, float64(message.Attempts))) * time.Second
		if retryErr := w.queue.Retry(ctx, message.ID, fmt.Sprintf("attempt %d: %v", message.Attempts+1, err), now.Add(delay)); retryErr != nil {
			return retryErr
		}
	}
	return nil
}
