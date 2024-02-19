package app

import (
	"fmt"
	"github.com/NuEventTeam/events/internal/handlers/http"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"log"
)

type App struct {
	httpServer *fiber.App
	port       int
}

func New(
	port int,
	httpHandler *http.Handler,
) *App {

	httpServer := fiber.New(fiber.Config{
		JSONEncoder: sonic.Marshal,
		JSONDecoder: sonic.Unmarshal,
	})

	httpServer.Use(cors.New(cors.Config{
		AllowHeaders:     "Origin,Content-Type,Authorization,Accept,Content-Length,Accept-Language,Accept-Encoding,Connection,Access-Control-Allow-Origin",
		AllowOrigins:     "*",
		AllowCredentials: true,
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
	}))

	httpHandler.SetUpRoutes(httpServer)

	return &App{
		httpServer: httpServer,
		port:       port,
	}
}

func (a *App) MustRun() {
	log.Fatal(a.httpServer.Listen(fmt.Sprintf("%d", a.port)))
}

func (a *App) Stop() {
	log.Fatal(a.httpServer.Shutdown())
}
