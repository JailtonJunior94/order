package bundle

import (
	"context"
	"database/sql"
	"log"

	"github.com/jailtonjunior94/order/configs"
	"github.com/jailtonjunior94/order/pkg/database/postgres"
)

type Container struct {
	DB     *sql.DB
	Config *configs.Config
}

func NewContainer(ctx context.Context) *Container {
	config, err := configs.LoadConfig(".")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := postgres.NewPostgresDatabase(config)
	if err != nil {
		log.Fatalf("failed to create database: %v", err)
	}

	return &Container{
		DB:     db,
		Config: config,
	}
}
