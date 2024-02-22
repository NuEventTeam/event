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
	"github.com/NuEventTeam/events/pkg"
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

	userSvc := users.NewEventSvc(db, cache)

	cdnService := cdn.New(pkg.CDNBaseUrl)

	httpHandler := http.NewHttpHandler(eventSvc, cdnService, userSvc, cfg.JWT.Secret)

	application := app.New(cfg.Http.Port, httpHandler)

	go application.MustRun()

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	application.Stop()

	log.Println("application stopped")
}
