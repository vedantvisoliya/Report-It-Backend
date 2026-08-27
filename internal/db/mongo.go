package db

import (
	"context"
	"fmt"
	"reportit-api/internal/config"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongo struct {
	Client *mongo.Client
	DB     *mongo.Database
}

func Connect(ctx context.Context, cfg config.Config) (*Mongo, error) {
	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(cfg.MongoURI)
	client, err := mongo.Connect(connectCtx, clientOpts)
	if err != nil {
		return &Mongo{}, fmt.Errorf("mongo connection failed (%w)", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return &Mongo{}, fmt.Errorf("mongo ping failed (%w)", err)
	}

	database := client.Database(cfg.MongoDB)

	return &Mongo{
		Client: client,
		DB:     database,
	}, nil
}

func Disconnet(ctx context.Context, client *mongo.Client) error {
	disconnetCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return client.Disconnect(disconnetCtx)
}
