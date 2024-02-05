package main

import (
	"context"
	"github.com/NuEventTeam/events/internal/app"
	"github.com/NuEventTeam/events/internal/config"
	"github.com/NuEventTeam/events/internal/handlers"
	event_service "github.com/NuEventTeam/events/internal/services/event"
	keydb "github.com/NuEventTeam/events/internal/storage/cache"
	"github.com/NuEventTeam/events/internal/storage/database"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	//config
	cfg := config.MustLoad()

	logger := log.New(os.Stdout, "INFO", log.Ldate|log.Ltime|log.Llongfile)

	logger.Println("starting application")

	db := database.NewDatabase(context.Background(), cfg.Database)
	//storage
	cache := keydb.New(context.Background(), cfg.Cache)

	//service
	eventSvc := event_service.NewEventSvc(db, cache)

	grpc := handlers.NewGRPCHandler(db, cache, eventSvc)

	application := app.New(logger, cfg.GRPC.Port, grpc)

	go application.MustRun()

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	application.Stop()

	logger.Println("application stopped")
}
