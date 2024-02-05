package main

import (
	"context"
	"github.com/NuEventTeam/events/internal/config"
	keydb "github.com/NuEventTeam/events/internal/storage/cache"
	"github.com/NuEventTeam/events/internal/storage/database"
)

func main() {
	//config
	cfg := config.MustLoad()

	db := database.NewDatabase(context.Background(), cfg.Database)
	//storage
	cache := keydb.New(context.Background(), cfg.Cache)

	//service

	//grpc

}
