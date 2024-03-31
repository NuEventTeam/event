package followers

import (
	"context"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/gofiber/fiber/v2"
	"strconv"
)

func Unfollow(db *database.Database) fiber.Handler {
	return func(ctx *fiber.Ctx) error {

		userId := ctx.Locals("userId").(int64)

		eventId, err := strconv.ParseInt(ctx.Params("eventId"), 10, 64)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "invalid event id", err)
		}

		err = removeFollower(ctx.Context(), db, eventId, userId)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "something went wrong", err)
		}

		return pkg.Success(ctx, nil)
	}
}

func removeFollower(ctx context.Context, db *database.Database, eventId, followerId int64) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	err = database.RemoveEventFollower(ctx, tx, eventId, followerId)
	if err != nil {
		return err
	}

	err = decreaseFollowerCount(ctx, tx, eventId)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func decreaseFollowerCount(ctx context.Context, db database.DBTX, userId int64) error {
	query := `update events set follower_count = follower_count - 1 where id = $1`

	_, err := db.Exec(ctx, query, userId)
	return err
}
