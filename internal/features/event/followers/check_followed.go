package followers

import (
	"context"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/gofiber/fiber/v2"
	"strconv"
)

func CheckIfFollowed(db *database.Database) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		userId := ctx.Locals("userId").(int64)

		eventId, err := strconv.ParseInt(ctx.Params("eventId"), 10, 64)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "invalid event id", err)
		}

		ok, err := CheckIfUserFollowed(ctx.Context(), db.GetDb(), eventId, userId)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "oops something went wrong", err)
		}

		return pkg.Success(ctx, fiber.Map{"follows": ok})
	}
}

func CheckIfUserFollowed(ctx context.Context, db database.DBTX, eventId, userId int64) (bool, error) {
	query := `select count(*) from event_followers where event_id = $1 and user_id = $2`

	args := []interface{}{eventId, userId}

	var count int64
	err := db.QueryRow(ctx, query, args...).Scan(&count)
	return count != 0, err
}
