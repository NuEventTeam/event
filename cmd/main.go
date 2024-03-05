package main

import (
	"context"
	"github.com/NuEventTeam/events/internal/app"
	"github.com/NuEventTeam/events/internal/config"
	"github.com/NuEventTeam/events/internal/handlers/http"
	"github.com/NuEventTeam/events/internal/services/cdn"
	event_service "github.com/NuEventTeam/events/internal/services/event"
	"github.com/NuEventTeam/events/internal/services/users"
	keydb "github.com/NuEventTeam/events/internal/storage/cache"
	"github.com/NuEventTeam/events/internal/storage/database"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const (
	localConfigPath = "./config/local.yaml"
	devConfigPath   = "./config/dev.yaml"
)

func main() {

	cfg := config.MustLoad(localConfigPath)

	db := database.NewDatabase(context.Background(), cfg.Database)

	cache := keydb.New(context.Background(), cfg.Cache)

	cdnService := cdn.New(config.CDNBaseUrl)

	userSvc := users.NewEventSvc(db, cache, cdnService)

	eventSvc := event_service.NewEventSvc(db, cache, cdnService)

	httpHandler := http.NewHttpHandler(eventSvc, cdnService, userSvc, cfg.JWT.Secret)

	application := app.New(cfg.Http.Port, httpHandler)

	go application.MustRun()

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	application.Stop()

	log.Println("application stopped")
}
