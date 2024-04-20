package user_profile

import (
	"context"
	"errors"
	sq "github.com/Masterminds/squirrel"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/NuEventTeam/events/pkg/types"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"time"
)

func GetLikedEventsHandler(db *database.Database) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		userId := ctx.Locals("userId").(int64)

		lastUserId := ctx.QueryInt("lastEventId", 0)

		followed, err := GetLikedEvents(ctx.Context(), db.GetDb(), userId, int64(lastUserId))
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, SomethingWentWrongMsg, err)
		}

		var events []FollowedEvent

		for _, val := range followed {
			events = append(events, val)
		}
		return pkg.Success(ctx, fiber.Map{"followedEvents": events})

	}
}

func GetLikedEvents(ctx context.Context, db database.DBTX, userId, lastEventId int64) (map[int64]FollowedEvent, error) {
	query := qb.Select(` 
					events.id,
					events.title, 
events.description,
					events.like_count,
					events.price,
					event_locations.address,
					event_locations.starts_at,
					event_locations.ends_at,
					event_locations.attendees_count`).
		From("events").
		InnerJoin("event_locations on event_locations.event_id = events.id").
		InnerJoin("event_like on event_like.event_id = events.id").
		Where(sq.Lt{"event_locations.starts_at": time.Now()}).Where(sq.Eq{"event_like.user_id": userId})

	if lastEventId != 0 {
		query = query.Where(sq.Lt{"events.id": lastEventId})
	}
	query = query.OrderBy("event_locations.starts_at desc")

	stmt, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(ctx, stmt, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	followedEventsMap := map[int64]FollowedEvent{}
	followedEventsSlice := []int64{}
	for rows.Next() {
		var (
			f FollowedEvent
			s time.Time
			e time.Time
		)

		err := rows.Scan(&f.ID, &f.Title, &f.Description, &f.LikesCount, &f.Price, &f.Address, &s, &e, &f.AttendeesCount)
		if err != nil {
			return nil, err
		}

		f.StartsAt = f.StartsAt.FromTime(&s)
		f.EndsAt = f.EndsAt.FromTime(&e)
		f.Date = types.Date(s)
		if f.Price != nil {
			*f.Price = *f.Price / 100
		}
		followedEventsMap[f.ID] = f
		followedEventsSlice = append(followedEventsSlice, f.ID)

	}
	err = getEventImages(ctx, db, followedEventsSlice, followedEventsMap)
	if err != nil {
		return nil, err
	}
	return followedEventsMap, nil

}
