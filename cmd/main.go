package main

import (
	"context"
	"github.com/NuEventTeam/events/internal/app"
	"github.com/NuEventTeam/events/internal/config"
	"github.com/NuEventTeam/events/internal/features/assets"
	"github.com/NuEventTeam/events/internal/features/auth"
	"github.com/NuEventTeam/events/internal/features/chat"
	"github.com/NuEventTeam/events/internal/features/event"
	"github.com/NuEventTeam/events/internal/features/handlers"
	"github.com/NuEventTeam/events/internal/features/sms_provider"
	"github.com/NuEventTeam/events/internal/features/user"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/internal/storage/keydb"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	cfg := config.MustLoad()

	db := database.NewDatabase(context.Background(), cfg.Database)
	cache := keydb.New(context.Background(), cfg.Cache)

	sms := sms_provider.New(cfg.SMS)
	assetsSvc := assets.NewS3Storage(cfg.CDN)

	userSvc := user.NewEventSvc(db, assetsSvc)

	eventSvc := event.NewEventSvc(db, assetsSvc)

	authSvc := auth.New(db, sms, cfg.JWT)

	httpHandler := handlers.New(eventSvc, cache, userSvc, assetsSvc, authSvc, cfg.JWT.Secret, db)

	application := app.New(cfg.Http.Port, httpHandler)

	go application.MustRun()

	go log.Println(chat.RunChatServer(cfg.Ws.Port, db))

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	application.Stop()
	log.Println("application stopped")
}
