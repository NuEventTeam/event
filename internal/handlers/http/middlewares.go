package http

import (
	"fmt"
	"github.com/NuEventTeam/events/pkg"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"strings"
)

func MustAuth(jwtSecret string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		tokenString := ctx.Get("Authorization")
		if tokenString == "" {
			return pkg.Error(ctx, http.StatusUnauthorized, "unauthorized", fmt.Errorf("invalid token"))
		}
		headerParts := strings.Split(tokenString, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			return pkg.Error(ctx, http.StatusUnauthorized, "unauthorized", fmt.Errorf("invalid token"))
		}
		userID, err := pkg.ParseJWT(headerParts[1], jwtSecret)
		if err != nil {
			return pkg.Error(ctx, http.StatusBadRequest, "invalid token", err)
		}
		ctx.Locals("user_id", userID)
		return ctx.Next()
	}
}
