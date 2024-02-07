package app

import (
	"fmt"
	"github.com/NuEventTeam/events/internal/handlers"
	"github.com/NuEventTeam/events/internal/storage/cache"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/protos/gen/go/event"
	"google.golang.org/grpc"
	"log"
	"log/slog"
	"net"
)

type App struct {
	log        *log.Logger
	gRPCServer *grpc.Server
	storage    *database.Database
	cache      *cache.Cache
	port       int
}

func New(
	log *log.Logger,
	port int,
	handler *handlers.GRPCHandler,
) *App {
	gRPCServer := grpc.NewServer()
	event.RegisterEventServiceServer(gRPCServer, handler)
	event.RegisterCategoriesServiceServer(gRPCServer, handler)
	event.RegisterUserServiceServer(gRPCServer, handler)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (a *App) MustRun() {

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))

	if err != nil {
		log.Fatal(err)
	}

	a.log.Println("handlers server is running", "addr", l.Addr().String())

	if err := a.gRPCServer.Serve(l); err != nil {
		log.Fatal(err)
	}
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.log.Println("stopping gRPC server", slog.Int("port", a.port))

	a.gRPCServer.GracefulStop()
}
