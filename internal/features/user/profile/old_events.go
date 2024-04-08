package user_profile

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/NuEventTeam/events/pkg/types"
	"github.com/gofiber/fiber/v2"
	"log"
	"time"
)

func GetOldEventsHandler(db *database.Database) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		userId := ctx.Locals("userId").(int64)

		lastUserId := ctx.QueryInt("lastEventId", 0)

		followed, ids, err := getOldEvents(ctx.Context(), db.GetDb(), userId, int64(lastUserId))
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, SomethingWentWrongMsg, err)
		}

		err = getEventImages(ctx.Context(), db.GetDb(), ids, followed)
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

func getOldEvents(ctx context.Context, db database.DBTX, userId, lastEventId int64) (map[int64]FollowedEvent, []int64, error) {
	query := qb.Select(` 
					events.id,
					events.title, 
					events.like_count,
					events.price,
					event_locations.address,
					event_locations.starts_at,
					event_locations.ends_at,
					event_locations.attendees_count`).
		From("events").
		InnerJoin("event_locations on event_locations.event_id = events.id").
		InnerJoin("event_followers on event_followers.event_id = events.id").
		Where(sq.LtOrEq{"event_locations.starts_at": time.Now()}).Where(sq.Eq{"event_followers.user_id": userId})

	if lastEventId != 0 {
		query = query.Where(sq.Lt{"events.id": lastEventId})
	}
	query = query.OrderBy("event_locations.starts_at desc")

	stmt, args, err := query.ToSql()
	if err != nil {
		return nil, nil, err
	}
	log.Println(stmt)
	rows, err := db.Query(ctx, stmt, args...)
	if err != nil {
		return nil, nil, err
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

		err := rows.Scan(&f.ID, &f.Title, &f.LikesCount, &f.Price, &f.Address, &s, &e, &f.AttendeesCount)
		if err != nil {
			return nil, nil, err
		}

		f.StartsAt = f.StartsAt.FromTime(&s)
		f.EndsAt = f.EndsAt.FromTime(&e)
		f.Date = types.Date(s)

		followedEventsMap[f.ID] = f
		followedEventsSlice = append(followedEventsSlice, f.ID)

	}
	return followedEventsMap, followedEventsSlice, nil

}
