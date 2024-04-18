package ticket

import (
	"context"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/internal/storage/keydb"
	"github.com/NuEventTeam/events/pkg"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/oklog/ulid/v2"
	"time"
)

func GenerateTicket(db *database.Database, cache *keydb.Cache, secret string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		userId := ctx.Locals("userId").(int64)

		eventId, err := ctx.ParamsInt("eventId")
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "invalid event id", err)
		}

		ok, err := checkIfFollows(ctx.Context(), db.GetDb(), int64(eventId), userId)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "oops something went wrong", err)
		}

		if !ok {
			return pkg.Error(ctx, fiber.StatusBadRequest, "user is not registered to event")
		}

		token, err := generateTicketToken(secret, int64(eventId), userId)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "oops something went wrong", err)
		}

		key := ulid.Make().String()

		err = cache.Set(ctx.Context(), key, token, time.Hour*24)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "oops something went wrong", err)
		}

		return pkg.Success(ctx, fiber.Map{
			"ticket": key,
		})
	}
}

func checkIfFollows(ctx context.Context, db database.DBTX, eventId, userId int64) (bool, error) {
	query := `select count(*) from events inner join event_followers on event_followers.event_id = events.id where status = 1 and events.id = $1 and event_followers.user_id = $2`

	args := []interface{}{eventId, userId}
	var count int64
	err := db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func generateTicketToken(secret string, eventId, userId int64) (string, error) {
	var (
		key = []byte(secret)
	)
	expireTime := time.Now().Add(time.Hour * 24)
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["userId"] = userId
	claims["eventId"] = eventId
	claims["exp"] = expireTime.Unix()

	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
