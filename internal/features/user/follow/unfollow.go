package user_follow

import (
	"context"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/gofiber/fiber/v2"
	"strconv"
)

var qb = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func UnfollowUser(db *database.Database) fiber.Handler {

	return func(ctx *fiber.Ctx) error {
		followerId := ctx.Locals("userId").(int64)
		userId, err := strconv.ParseInt(ctx.Params("userId"), 10, 64)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "invalid follower id", err)
		}

		err = RemoveFollower(ctx.Context(), db, userId, followerId)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, "something went wrong", err)
		}
		return pkg.Success(ctx, nil)
	}
}

func RemoveFollower(ctx context.Context, db *database.Database, userId, followerId int64) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	err = RemoveUserFollower(ctx, tx, userId, followerId)
	if err != nil {
		return err
	}
	err = DecreaseFollowerCount(ctx, tx, userId)
	if err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func RemoveUserFollower(ctx context.Context, db database.DBTX, userId, followerId int64) error {
	query := qb.Delete("user_followers").
		Where(sq.Eq{"user_id": userId}).
		Where(sq.Eq{"follower_id": followerId})

	stmt, args, err := query.ToSql()
	if err != nil {
		return err
	}

	row, err := db.Exec(ctx, stmt, args...)
	if row.RowsAffected() == 0 {
		return fmt.Errorf("follower does not exist")
	}
	return err
}
func DecreaseFollowerCount(ctx context.Context, db database.DBTX, userId int64) error {
	query := `update users set follower_count = follower_count - 1 where id = $1`

	_, err := db.Exec(ctx, query, userId)
	return err
}
