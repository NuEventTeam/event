package main

import (
	"context"
	"github.com/NuEventTeam/events/internal/app"
	"github.com/NuEventTeam/events/internal/config"
	"github.com/NuEventTeam/events/internal/features/event"
	"github.com/NuEventTeam/events/internal/features/user"
	"github.com/NuEventTeam/events/internal/handlers/http"
	"github.com/NuEventTeam/events/internal/services/cdn"
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

	cdnService := cdn.New(pkg.CDNBaseUrl)

	userSvc := user.NewEventSvc(db, cache, cdnService)

	eventSvc := event.NewEventSvc(db, cache, cdnService)

	httpHandler := http.NewHttpHandler(eventSvc, cdnService, userSvc, cfg.JWT.Secret)

	application := app.New(cfg.Http.Port, httpHandler)

	go application.MustRun()

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	application.Stop()

	log.Println("application stopped")
}
