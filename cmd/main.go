package main

import (
	"context"
	"github.com/NuEventTeam/events/internal/app"
	"github.com/NuEventTeam/events/internal/config"
	"github.com/NuEventTeam/events/internal/handlers/http"
	event_service "github.com/NuEventTeam/events/internal/services/event"
	keydb "github.com/NuEventTeam/events/internal/storage/cache"
	"github.com/NuEventTeam/events/internal/storage/database"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	cfg := config.MustLoad()

	db := database.NewDatabase(context.Background(), cfg.Database)

	cache := keydb.New(context.Background(), cfg.Cache)

	eventSvc := event_service.NewEventSvc(db, cache)

	httpHandler := http.NewHttpHandler(eventSvc, cfg.JWT.Secret)

	application := app.New(cfg.Http.Port, httpHandler)

	go application.MustRun()

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	application.Stop()

	log.Println("application stopped")
}
