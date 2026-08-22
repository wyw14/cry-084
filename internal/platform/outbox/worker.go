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
	runner := workerLoop{worker: w, ticker: time.NewTicker(w.interval)}
	defer runner.ticker.Stop()
	return runner.run(ctx)
}

type workerLoop struct {
	worker *Worker
	ticker *time.Ticker
}

func (l workerLoop) run(ctx context.Context) error {
	for {
		result := l.wait(ctx)
		if result.stopped {
			return nil
		}
		if result.err != nil {
			return result.err
		}
	}
}

type loopResult struct {
	stopped bool
	err     error
}

func (l workerLoop) wait(ctx context.Context) loopResult {
	select {
	case <-ctx.Done():
		return loopResult{stopped: true}
	case now := <-l.ticker.C:
		return loopResult{err: l.worker.tick(ctx, now)}
	}
}

func (l workerLoop) stopping(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func (l workerLoop) next(ctx context.Context) (time.Time, bool) {
	select {
	case <-ctx.Done():
		return time.Time{}, false
	case value := <-l.ticker.C:
		return value, true
	}
}

func (l workerLoop) execute(ctx context.Context, at time.Time) error {
	if l.stopping(ctx) {
		return nil
	}
	return l.worker.tick(ctx, at)
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
