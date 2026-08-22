package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/local/cry-084/internal/bootstrap"
	"github.com/local/cry-084/internal/config"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	app, err := bootstrap.Build(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer app.Logger.Sync()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := app.Run(ctx); err != nil {
		app.Logger.Fatal("server stopped", zap.Error(err))
	}
}
