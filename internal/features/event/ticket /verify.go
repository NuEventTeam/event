package ticket

import (
	"context"
	"errors"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/internal/storage/keydb"
	"github.com/NuEventTeam/events/pkg"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func VerifyTicket(db *database.Database, cache *keydb.Cache, secret string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		ticketId := ctx.Query("ticket")

		ticket, err := cache.Get(ctx.Context(), ticketId)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
		}

		followerId, eventId, err := ParseTicket(ticket.(string), secret)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
		}

		ok, err := checkIfFollows(ctx.Context(), db.GetDb(), eventId, followerId)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
		}
		if !ok {
			return pkg.Error(ctx, fiber.StatusBadRequest, "user is not registered to event")
		}
		err = setAttended(ctx.Context(), db.GetDb(), eventId, followerId)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
		}

		return pkg.Success(ctx, nil)
	}
}

func setAttended(ctx context.Context, db database.DBTX, eventId, userId int64) error {
	query := `update event_followers set atteded = true , updated_at = now() where event_id = $1 and user_id = $2`

	args := []interface{}{eventId, userId}

	_, err := db.Exec(ctx, query, args...)
	return err
}

func ParseTicket(ticket string, secret string) (int64, int64, error) {
	token, err := jwt.Parse(ticket, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Invalid sign method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return 0, 0, err
	}

	if !token.Valid {
		return 0, 0, jwt.ErrTokenExpired
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, 0, jwt.ErrTokenInvalidClaims
	}
	userId := int64(claims["userId"].(float64))
	eventId := int64(claims["eventId"].(float64))

	return userId, eventId, nil
}
