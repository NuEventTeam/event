package event

import (
	"context"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/NuEventTeam/events/internal/features/assets"
	"github.com/NuEventTeam/events/internal/models"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"log"
	"time"
)

func (e Event) GetEventByIDHandler() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		eventID, err := ctx.ParamsInt("eventId", 0)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, "not proper id", err)

		}
		event, err := e.getEventByID(ctx.Context(), int64(eventID))
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
		}

		return pkg.Success(ctx, event)
	}
}
func (e Event) GetAllEvenst() fiber.Handler {
	return func(ctx *fiber.Ctx) error {

		lastId := ctx.QueryInt("lastId", 0)
		eventsMap, err := e.getEventAllEvents(ctx.Context(), e.db.GetDb(), int64(lastId))
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
		}
		events := []models.Event{}

		for _, val := range eventsMap {
			events = append(events, val)
		}
		return pkg.Success(ctx, events)
	}
}
func (e *Event) getEventByID(ctx context.Context, eventId int64) (*models.Event, error) {
	event, err := database.GetEventByID(ctx, e.db.GetDb(), eventId)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, fmt.Errorf("event not found")
	}

	categories, err := database.GetEventCategories(ctx, e.db.GetDb(), eventId)
	if err != nil {
		return nil, err
	}
	locations, err := database.GetEventLocations(ctx, e.db.GetDb(), eventId)
	if err != nil {
		return nil, err
	}

	images, err := database.GetEventImages(ctx, e.db.GetDb(), eventId)
	if err != nil {
		return nil, err
	}

	managers, err := database.GetEventManagers(ctx, e.db.GetDb(), eventId)
	if err != nil {
		return nil, err
	}

	event.Categories = categories
	event.Locations = locations
	event.Images = images
	event.Managers = managers

	return event, nil
}

func (e *Event) getEventAllEvents(ctx context.Context, db database.DBTX, lastId int64) (map[int64]models.Event, error) {
	query := qb.Select("id", "title", "description", "age_min", "age_max", "status", "created_at", "follower_count", "price").
		From("events")

	if lastId > 0 {
		query = query.Where(sq.Gt{"id": lastId})
	}

	query = query.Limit(15)
	stmt, params, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	log.Println(stmt)
	events := map[int64]models.Event{}
	var price *int64
	rows, err := db.Query(ctx, stmt, params...)
	if err != nil {
		return nil, err
	}
	var eventIds []int64
	for rows.Next() {
		var event models.Event
		err := rows.Scan(&event.ID, &event.Title, &event.Description, &event.MinAge, &event.MaxAge, &event.Status, &event.CreatedAt, &event.FollowerCount, &price)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, nil
			}
			return nil, err
		}
		if price != nil {
			event.Price = float64(*price) / 100
		}
		events[event.ID] = event
		eventIds = append(eventIds, event.ID)
	}
	rows.Close()
	query = qb.Select("event_id,category_id", "categories.name").
		From("event_categories").
		InnerJoin("categories on event_categories.category_id = categories.id").
		Where(sq.Eq{"event_id": eventIds})

	stmt, params, err = query.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err = db.Query(ctx, stmt, params...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var (
			eventId int64
			c       models.Category
		)

		err := rows.Scan(&eventId, &c.ID, &c.Name)
		if err != nil {
			return nil, err
		}

		val := events[eventId]
		val.Categories = append(val.Categories, c)
		events[eventId] = val
	}
	rows.Close()

	query = qb.Select("id", "event_id", "address", "longitude", "latitude", "seats", "starts_at", "ends_at", "attendees_count").
		From("event_locations").
		Where(sq.Eq{"deleted_at": nil}).
		Where(sq.Eq{"event_id": eventIds})

	stmt, params, err = query.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err = db.Query(ctx, stmt, params...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var (
			l models.Location
			s time.Time
			e time.Time
		)

		err := rows.Scan(&l.ID, &l.EventID, &l.Address, &l.Longitude, &l.Latitude, &l.Seats, &s, &e, &l.AttendeesCount)
		if err != nil {
			return nil, err
		}

		l.StartsAt = l.StartsAt.FromTime(&s)
		l.EndsAt = l.EndsAt.FromTime(&e)
		val := events[l.EventID]
		val.Locations = append(val.Locations, l)
		events[l.EventID] = val
	}

	rows.Close()

	query = qb.Select("id", "event_id", "url").
		From("event_images").
		Where(sq.Eq{"deleted_at": nil}).
		Where(sq.Eq{"event_id": eventIds})

	stmt, params, err = query.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err = db.Query(ctx, stmt, params...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var i assets.Image

		err := rows.Scan(&i.ID, &i.EventID, &i.Url)
		if err != nil {
			return nil, err
		}
		i.Url = pkg.CDNBaseUrl + i.Url
		val := events[i.EventID]
		val.Images = append(val.Images, i)
		events[i.EventID] = val
	}

	query = qb.Select(
		"users.username",
		"users.firstname",
		"users.lastname",
		"users.profile_image",
		"event_managers.user_id",
		"event_managers.role_id",
		"event_roles.name",
		"users.phone",
		"event_managers.event_id",
	).
		From("users").
		InnerJoin("event_managers on event_managers.user_id = users.id").
		InnerJoin("event_roles on event_roles.id = event_managers.role_id").
		Where(sq.Eq{"event_managers.event_id": eventIds}).
		Where(sq.Eq{"event_managers.deleted_at": nil})

	stmt, params, err = query.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err = db.Query(ctx, stmt, params...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var m models.Manager

		err := rows.Scan(&m.User.Username, &m.User.Firstname, &m.User.Lastname, &m.User.ProfileImage, &m.User.UserID, &m.Role.ID, &m.Role.Name, &m.User.Phone, &m.EventId)
		if err != nil {
			return nil, err
		}
		if m.User.ProfileImage != nil {
			*m.User.ProfileImage = pkg.CDNBaseUrl + "/" + *m.User.ProfileImage
		}

		query := qb.Select("permission_id").
			From("event_role_permissions").
			Where(sq.Eq{"role_id": m.Role.ID})

		stmt, params, err := query.ToSql()
		if err != nil {
			return nil, err
		}

		rows, err := db.Query(ctx, stmt, params...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var p int64
			err := rows.Scan(&p)
			if err != nil {
				return nil, err
			}

			m.Role.Permissions = append(m.Role.Permissions, p)

		}

		val := events[m.EventId]
		val.Managers = append(val.Managers, m)
		events[m.EventId] = val
	}

	defer rows.Close()

	return events, nil
}
