package user_follow

import (
	"context"
	"errors"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"log"
	"strconv"
)

func FollowUser(db *database.Database) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		followerId := ctx.Locals("userId").(int64)
		userId, err := strconv.ParseInt(ctx.Params("userId"), 10, 64)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "invalid follower id", err)
		}
		log.Println(followerId, userId)
		err = AddFollower(ctx.Context(), db, userId, followerId)
		if err != nil {
			log.Println(err)
			return pkg.Error(ctx, fiber.StatusInternalServerError, "something went wrong", err)
		}
		return pkg.Success(ctx, nil)
	}
}

func AddFollower(ctx context.Context, db *database.Database, userId, followerId int64) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	err = database.AddUserFollower(ctx, tx, userId, followerId)
	if err != nil {
		var e *pgconn.PgError
		if errors.As(err, &e) && e.Code == pgerrcode.UniqueViolation {
			return nil
		}
		return err
	}

	err = IncreaseFollowerCount(ctx, tx, userId)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func IncreaseFollowerCount(ctx context.Context, db database.DBTX, userId int64) error {
	query := `update users set follower_count = follower_count + 1 where id = $1`

	_, err := db.Exec(ctx, query, userId)
	return err
}
