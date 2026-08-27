package app

import (
	"context"
	"fmt"
	"reportit-api/internal/config"
	"reportit-api/internal/db"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

type App struct {
	Config config.Config
	Client *mongo.Client
	DB     *mongo.Database
}

func StartApp(ctx context.Context) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return &App{}, err
	}

	mongoClient, err := db.Connect(ctx, cfg)
	if err != nil {
		return &App{}, err
	}

	return &App{
		Config: cfg,
		Client: mongoClient.Client,
		DB:     mongoClient.DB,
	}, nil
}

func (a *App) CloseApp(ctx context.Context) error {
	if a.Client == nil {
		return nil
	}

	closeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.Disconnet(closeCtx, a.Client); err != nil {
		return fmt.Errorf("mongo disconnect failed (%w)", err)
	}

	return nil
}
