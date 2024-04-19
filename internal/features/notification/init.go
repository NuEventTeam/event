package notification

import (
	"context"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"github.com/NuEventTeam/events/pkg"
	"github.com/gofiber/fiber/v2"
	"log"
)

var notify *messaging.Client

func New(app *firebase.App) {
	client, err := app.Messaging(context.Background())
	if err != nil {
		log.Fatalf("error getting Messaging client: %v\n", err)
	}

	notify = client

}

func SetUpNotification(router *fiber.App) {
	router.Get("sms", TestNotification())
}

func TestNotification() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		token := ctx.Query("token")
		err := Send([]string{token}, map[string]string{"text": "hello"})
		log.Println(err)
		return pkg.Success(ctx, nil)
	}
}

func Send(tokens []string, data map[string]string) error {
	message := &messaging.MulticastMessage{
		Data:   data,
		Tokens: tokens,
	}

	_, err := notify.SendMulticast(context.Background(), message)
	return err
}
