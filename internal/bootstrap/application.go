package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"time"

	hazardapp "github.com/local/cry-084/internal/application/hazard"
	inspectionapp "github.com/local/cry-084/internal/application/inspection"
	"github.com/local/cry-084/internal/config"
	"github.com/local/cry-084/internal/platform/clock"
	"github.com/local/cry-084/internal/platform/identity"
	"github.com/local/cry-084/internal/platform/outbox"
	"github.com/local/cry-084/internal/repository/memory"
	httptransport "github.com/local/cry-084/internal/transport/http"
	"go.uber.org/zap"
)

type Application struct {
	Server *http.Server
	Logger *zap.Logger
}

func Build(cfg config.Config) (Application, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return Application{}, err
	}
	repo := memory.New()
	systemClock := clock.System{}
	ids := identity.Random{}
	queue := outbox.NewMemory()
	inspections := inspectionapp.NewService(repo, systemClock, ids)
	hazards := hazardapp.NewService(repo, systemClock, ids, queue)
	handler := httptransport.NewHandler(inspections, hazards)
	router := httptransport.NewRouter(handler, logger, cfg.RequestTimeout)
	server := &http.Server{Addr: cfg.Address, Handler: router, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	return Application{Server: server, Logger: logger}, nil
}
func (a Application) Run(ctx context.Context) error {
	errors := make(chan error, 1)
	go func() { errors <- a.Server.ListenAndServe() }()
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := a.Server.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return nil
	case err := <-errors:
		if err == http.ErrServerClosed {
			return nil
		}
		return fmt.Errorf("http server: %w", err)
	}
}
