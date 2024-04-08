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

var qb = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

var SomethingWentWrongMsg = "oops something went wrong"

func GetFollowedEventsHandler(db *database.Database) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		userId := ctx.Locals("userId").(int64)

		lastUserId := ctx.QueryInt("lastEventId", 0)
		followed, ids, err := getFollowedEvents(ctx.Context(), db.GetDb(), userId, int64(lastUserId))
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

type FollowedEvent struct {
	ID             int64           `json:"id"`
	Title          string          `json:"title"`
	Address        string          `json:"address"`
	Date           types.Date      `json:"date"`
	StartsAt       *types.DateTime `json:"startsAt"`
	EndsAt         *types.DateTime `json:"endsAt"`
	Images         []string        `json:"images"`
	Price          *float64        `json:"price"`
	AttendeesCount int64           `json:"attendeesCount"`
	LikesCount     int64           `json:"likesCount"`
}

func getFollowedEvents(ctx context.Context, db database.DBTX, userId, lastEventId int64) (map[int64]FollowedEvent, []int64, error) {
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
		Where(sq.GtOrEq{"event_locations.starts_at": time.Now()}).
		Where(sq.Eq{"event_followers.user_id": userId})

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
		log.Println("here")
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
		if f.Price != nil {
			*f.Price = *f.Price / 100
		}

		followedEventsMap[f.ID] = f
		followedEventsSlice = append(followedEventsSlice, f.ID)

	}
	return followedEventsMap, followedEventsSlice, nil

}

func getEventImages(ctx context.Context, db database.DBTX, eventIds []int64, events map[int64]FollowedEvent) error {
	query := qb.Select("event_id, url").From("event_images").Where(sq.Eq{"event_id": eventIds})

	stmt, args, err := query.ToSql()
	if err != nil {
		return err
	}

	rows, err := db.Query(ctx, stmt, args...)
	if err != nil {
		log.Println("here")
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var (
			eventId int64
			url     *string
		)
		err := rows.Scan(&eventId, &url)
		if err != nil {
			return err
		}
		if url != nil {
			*url = pkg.CDNBaseUrl + "/get/" + *url
			val, _ := events[eventId]
			val.Images = append(events[eventId].Images, *url)
			events[eventId] = val
		}
	}
	return nil
}
