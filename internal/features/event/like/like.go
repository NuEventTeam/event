package like

import (
	"context"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/gofiber/fiber/v2"
)

func LikeEvent(db *database.Database) fiber.Handler {
	return func(ctx *fiber.Ctx) error {

		userId := ctx.Locals("userId").(int64)

		eventId, err := ctx.ParamsInt("eventId")
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
		}

		liked, err := isLiked(ctx.Context(), db.GetDb(), int64(eventId), userId)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, "oops something went wrong", err)
		}

		if liked {
			err := addLike(ctx.Context(), db, int64(eventId), userId)
			if err != nil {
				return pkg.Error(ctx, fiber.StatusInternalServerError, "oops something went wrong", err)
			}
		} else {
			err := removeLike(ctx.Context(), db, int64(eventId), userId)
			if err != nil {
				return pkg.Error(ctx, fiber.StatusInternalServerError, "oops something went wrong", err)
			}
		}

		return pkg.Success(ctx, nil)
	}
}

func isLiked(ctx context.Context, db database.DBTX, eventId, userId int64) (bool, error) {
	query := `select count(*) from event_like where event_id = $1 and user_id = $2`

	var count int64

	err := db.QueryRow(ctx, query, eventId, userId).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func addLike(ctx context.Context, db *database.Database, eventId, userId int64) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `insert into  event_like (event_id, user_id) values ($1, $2)`

	_, err = tx.Exec(ctx, query, eventId, userId)
	if err != nil {
		return err
	}

	query = `update events
				set like_count = like_count + 1
				where id = $1
	`
	_, err = tx.Exec(ctx, query, eventId)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func removeLike(ctx context.Context, db *database.Database, eventId, userId int64) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `delete from event_like where event_id = $1 and user_id = $2`

	_, err = tx.Exec(ctx, query, eventId, userId)
	if err != nil {
		return err
	}

	query = `update events
				set like_count = like_count - 1
				where id = $1
	`
	_, err = tx.Exec(ctx, query, eventId)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}
