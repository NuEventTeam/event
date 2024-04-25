package handlers

import (
	"errors"
	"fmt"
	"github.com/NuEventTeam/events/internal/features/event"
	"github.com/NuEventTeam/events/pkg"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"strconv"
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
		userID, userAgent, err := pkg.ParseJWT(headerParts[1], jwtSecret)
		if err != nil {
			return pkg.Error(ctx, http.StatusBadRequest, "invalid token", err)
		}
		ctx.Locals("userId", userID)
		ctx.Locals("userAgent", userAgent)
		return ctx.Next()
	}
}

func ExtractUserIdFromAuthHeader(jwtSecret string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		tokenString := ctx.Get("Authorization")
		if tokenString == "" {
			ctx.Locals("userId", int64(0))
			return ctx.Next()
		}
		headerParts := strings.Split(tokenString, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			ctx.Locals("userId", int64(0))
		}
		userID, userAgent, err := pkg.ParseJWT(headerParts[1], jwtSecret)
		if err != nil {
			ctx.Locals("userId", int64(0))
		}
		ctx.Locals("userId", userID)
		ctx.Locals("userAgent", userAgent)
		return ctx.Next()
	}
}

func (h *Handler) HasPermission(permission int64) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		userId := ctx.Locals("userId").(int64)
		eventId, err := strconv.ParseInt(ctx.Params("eventId"), 10, 64)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "invalid eventID", err)
		}

		err = h.EventSvc.CheckPermission(ctx.Context(), eventId, userId, permission)
		if err != nil {
			if errors.Is(err, event.ErrNoPermission) {
				return pkg.Error(ctx, fiber.StatusForbidden, "has no permission")
			}
			return pkg.Error(ctx, fiber.StatusInternalServerError, "something went wrong", err)
		}

		return ctx.Next()
	}
}
