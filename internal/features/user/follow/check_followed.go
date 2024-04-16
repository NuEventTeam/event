package user_follow

import (
	"context"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/gofiber/fiber/v2"
)

func CheckFollowed(db *database.Database) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		followedId, err := ctx.ParamsInt("userId")
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "invalid user id", err)
		}

		ok, err := checkFollowed(ctx.Context(), db.GetDb(), ctx.Locals("userId").(int64), int64(followedId))
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "oops something went wrong", err)
		}
		return pkg.Success(ctx, fiber.Map{"followed": ok})
	}
}

func checkFollowed(ctx context.Context, db database.DBTX, userId, followedId int64) (bool, error) {
	query := `select count(*) from user_followers where follower_id=$1 and user_id=$2`

	var count int64

	err := db.QueryRow(ctx, query, userId, followedId).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil

}
