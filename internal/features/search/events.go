package search

import (
	"context"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/NuEventTeam/events/pkg/types"
	"github.com/gofiber/fiber/v2"
	"time"
)

type Sort struct {
	By    string `json:"by"`
	Order string `json:"order"`
}

type Coordinate struct {
	MaxLon    float64 `json:"maxLon"`
	MinLon    float64 `json:"minLon"`
	MaxLat    float64 `json:"maxLat"`
	MinLat    float64 `json:"minLat"`
	CenterLat float64 `json:"centerLat"`
	CenterLog float64 `json:"centerLog"`
}

type SearchArgs struct {
	Text       string     `json:"text"`
	Coordinate Coordinate `json:"coordinate"`
	Categories []int64    `json:"categories"`
	MinAge     int64      `json:"minAge"`
	Sort       []Sort     `json:"sort"`
	LastId     int64      `json:"lastId"`
}

func SearchEvents(db *database.Database) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var args SearchArgs

		if err := ctx.BodyParser(&args); err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
		}

		eventsMap, eventIds, err := searchForEvent(ctx.Context(), db.GetDb(), args)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, "oops something went wrong", err)
		}

		categoies, err := getCategories(ctx.Context(), db.GetDb(), eventIds)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, "oops something went wrong", err)
		}

		var events []Event

		for id, event := range eventsMap {
			event.Categories = categoies[id]
			events = append(events, event)
		}

		return pkg.Success(ctx, fiber.Map{"events": events})

	}
}

type Location struct {
	Address        string `json:"address"`
	Log            string `json:"lon"`
	Lat            string `json:"lat"`
	AttendeesCount int64  `json:"attendeesCount"`
	Seats          int64  `json:"seats"`
}

type Event struct {
	Id            int64           `json:"eventId"`
	Title         string          `json:"title"`
	Location      Location        `json:"location"`
	Images        []string        `json:"images"`
	Categories    []Categories    `json:"categories"`
	Author        User            `json:"author"`
	AgeMin        *int64          `json:"ageMin"`
	StartsAt      *types.DateTime `json:"startsAt"`
	EndsAt        *types.DateTime `json:"endsAt"`
	Distance      float64         `json:"distance"`
	LikeCount     int64           `json:"likeCount"`
	FollowerCount int64           `json:"followerCount"`
}

func searchForEvent(ctx context.Context, db database.DBTX, params SearchArgs) (map[int64]Event, []int64, error) {
	query := qb.Select("events.id,title,age_min, like_count, follower_count, username,user_id,url" +
		"address, longitude, latitude, seats, attendees_count, starts_at, ends_at, distance as").
		From("events").
		InnerJoin("events_locations on events.id = event_locations.event_id").
		InnerJoin("event_managers on events.id = event_managers.event_id").
		InnerJoin("event_role_permissions on event_managers.role_id = event_role_permissions.role_id").
		InnerJoin("event_categories on events.id = event_categories.event_id").
		Where(sq.Eq{"permission_id": pkg.PermissionUpdate}).
		Where(sq.And{
			sq.GtOrEq{"latitude": params.Coordinate.MinLat},
			sq.LtOrEq{"latitude": params.Coordinate.MaxLat}}).
		Where(sq.And{
			sq.GtOrEq{"longitude": params.Coordinate.MinLon},
			sq.LtOrEq{"longitude": params.Coordinate.MaxLat}}).
		Where(sq.GtOrEq{"age_min": params.MinAge}).Where(sq.Gt{"events.id": params.LastId})

	if len(params.Categories) > 0 {
		query = query.Where(sq.Eq{"event_categories.event_id": params.Categories})
	}

	query = query.Where(sq.Like{"events.title": "%" + params.Text + "%"})
	query = query.Where(sq.Like{"events.description": "%" + params.Text + "%"})

	query = query.Column(`2 * asin(sqrt(
		pow(sin(radians(`+sq.Placeholders(1)+`- latitude) / 2), 2) +
		cos(radians(`+sq.Placeholders(1)+`)) * cos(radians(latitude)) * pow(sin(radians(`+sq.Placeholders(1)+`- longitude) / 2), 2)
)) * 6371 AS distance_in_km`, params.Coordinate.CenterLat, params.Coordinate.CenterLat, params.Coordinate.CenterLog)

	for _, val := range params.Sort {
		query = query.OrderBy(fmt.Sprintf("%s %s", val.By, val.By))
	}

	stmt, args, err := query.ToSql()
	if err != nil {
		return nil, nil, err
	}

	rows, err := db.Query(ctx, stmt, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	events := map[int64]Event{}
	eventIds := []int64{}
	for rows.Next() {
		var (
			e        Event
			startsAt time.Time
			endsAt   time.Time
		)

		err := rows.Scan(&e.Id, &e.Title, &e.AgeMin, &e.LikeCount, &e.FollowerCount, &e.Author.ID, &e.Author.Username, &e.Author.ProfileImage,
			&e.Location.Address, &e.Location.Log, &e.Location.Lat, &e.Location.Seats, &e.Location.AttendeesCount, &startsAt, &endsAt,
			&e.Distance)
		if err != nil {
			return nil, nil, err
		}

		if !startsAt.IsZero() {
			e.StartsAt = e.StartsAt.FromTime(&startsAt)
		}
		if !endsAt.IsZero() {
			e.EndsAt = e.EndsAt.FromTime(&endsAt)
		}

		events[e.Id] = e
		eventIds = append(eventIds, e.Id)

	}
	return events, eventIds, err
}

func getEventCategories(ctx context.Context, db database.DBTX, eventIds []int64) (map[int64][]Categories, error) {
	query := qb.Select("categories.id, categories.name,event_categories.event_id").
		From("categories").
		InnerJoin("event_categories on event_categories.category_id = categories.id").
		Where(sq.Eq{"event_categories.event_id": eventIds})

	stmt, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		categories map[int64][]Categories
	)

	for rows.Next() {
		var (
			c       Categories
			eventId int64
		)

		err := rows.Scan(&c.ID, &c.Name, &eventId)
		if err != nil {
			return nil, err
		}

		if _, ok := categories[eventId]; !ok {
			categories[eventId] = []Categories{}
		}
		categories[eventId] = append(categories[eventId], c)

	}
	return categories, nil

}
