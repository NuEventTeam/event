package user

import (
	"context"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/gofiber/fiber/v2"
	"strconv"
)

func (u User) UnfollowUser() fiber.Handler {

	return func(ctx *fiber.Ctx) error {
		followerId := ctx.Locals("userId").(int64)
		userId, err := strconv.ParseInt(ctx.Params("userId"), 10, 64)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "invalid follower id", err)
		}

		err = u.RemoveFollower(ctx.Context(), userId, followerId)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, "something went wrong", err)
		}
		return pkg.Success(ctx, nil)
	}
}

func (e *User) RemoveFollower(ctx context.Context, userId, followerId int64) error {
	tx, err := e.db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	err = database.RemoveUserFollower(ctx, tx, userId, followerId)
	if err != nil {
		return err
	}
	err = database.UpdateUserFollowerCount(ctx, tx, userId, -1)
	if err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}
