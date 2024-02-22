package database

import (
	"context"
	"errors"
	sq "github.com/Masterminds/squirrel"
	"github.com/NuEventTeam/events/internal/models"
	"github.com/NuEventTeam/events/pkg"
	"github.com/jackc/pgx/v5"
	"log"
	"strconv"
	"time"
)

func CreateEvent(ctx context.Context, db DBTX, event models.Event) (int64, error) {
	query := qb.Insert("events").
		Columns("title", "description", "age_min", "age_max").
		Values(event.Title, event.Description, event.MinAge, event.MaxAge).
		Suffix("RETURNING id")

	stmt, params, err := query.ToSql()
	if err != nil {
		return 0, err
	}

	var id int64

	err = db.QueryRow(ctx, stmt, params...).Scan(&id)

	return id, err
}

func AddEventManager(ctx context.Context, db DBTX, eventID int64, manager ...models.Manager) error {
	if len(manager) == 0 {
		return nil
	}
	query := qb.Insert("event_managers").
		Columns("event_id", "user_id", "role_id")

	for _, m := range manager {
		query = query.Values(eventID, m.User.UserID, m.Role.ID)
	}

	stmt, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, stmt, args...)
	return err
}

func AddEventCategories(ctx context.Context, db DBTX, eventID int64, category ...int64) error {
	if len(category) == 0 {
		return nil
	}
	query := qb.Insert("event_categories").
		Columns("event_id", "category_id")

	for _, id := range category {
		query = query.Values(eventID, id)
	}

	stmt, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, stmt, args...)
	return err
}

func AddEventLocations(ctx context.Context, db DBTX, eventID int64, locations ...models.Location) error {

	if len(locations) == 0 {
		return nil
	}
	query := qb.Insert("event_locations").
		Columns("event_id", "address", "longitude", "latitude", "seats", "starts_at", "ends_at")

	for _, l := range locations {
		lg := strconv.FormatFloat(*l.Longitude, 'f', 12, 32)
		lt := strconv.FormatFloat(*l.Latitude, 'f', 12, 32)
		log.Println(lg, lt)
		query = query.Values(eventID, l.Address, lg, lt, l.Seats, l.StartsAt, l.EndsAt)
	}

	stmt, params, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, stmt, params...)

	return err

}

func AddEventImage(ctx context.Context, db DBTX, eventID int64, images ...models.Image) error {
	if len(images) == 0 {
		return nil
	}

	query := qb.Insert("event_images").
		Columns("event_id", "url")

	for _, i := range images {
		query = query.Values(eventID, i.Url)
	}

	stmt, params, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, stmt, params...)

	return err

}

func GetEventByID(ctx context.Context, db DBTX, eventID int64) (*models.Event, error) {
	query := qb.Select("id", "title", "description", "age_min", "age_max", "created_at").
		From("events").Where(sq.Eq{"id": eventID})

	stmt, params, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	log.Println(stmt)
	event := &models.Event{}

	err = db.QueryRow(ctx, stmt, params...).Scan(&event.ID, &event.Title, &event.Description, &event.MinAge, &event.MaxAge, &event.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return event, nil
}

func GetEventLocations(ctx context.Context, db DBTX, eventID int64) ([]models.Location, error) {
	query := qb.Select("id", "event_id", "address", "longitude", "latitude", "seats", "starts_at", "ends_at").
		From("event_locations").
		Where(sq.Eq{"deleted_at": nil}).
		Where(sq.Eq{"event_id": eventID})

	stmt, params, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var locs []models.Location

	rows, err := db.Query(ctx, stmt, params...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var l models.Location
		err := rows.Scan(&l.ID, &l.EventID, &l.Address, &l.Longitude, &l.Latitude, &l.Seats, &l.StartsAt, &l.EndsAt)
		if err != nil {
			return nil, err
		}
		locs = append(locs, l)
	}

	return locs, nil
}

func GetEventImages(ctx context.Context, db DBTX, eventId int64, imgIds ...int64) ([]models.Image, error) {
	query := qb.Select("id", "event_id", "url").
		From("event_images").
		Where(sq.Eq{"deleted_at": nil}).
		Where(sq.Eq{"event_id": eventId})
	if len(imgIds) > 0 {
		query = query.Where(sq.Eq{"id": imgIds})
	}

	stmt, params, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var imgs []models.Image

	rows, err := db.Query(ctx, stmt, params...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var i models.Image

		err := rows.Scan(&i.ID, &i.EventID, &i.Url)
		if err != nil {
			return nil, err
		}
		i.Url = pkg.CDNBaseUrl + "/get/" + i.Url
		imgs = append(imgs, i)
	}
	return imgs, nil
}

func GetEventCategories(ctx context.Context, db DBTX, eventId int64) ([]models.Category, error) {
	query := qb.Select("category_id", "categories.name").
		From("event_categories").
		InnerJoin("categories on event_categories.category_id = categories.id").
		Where(sq.Eq{"event_id": eventId})

	stmt, params, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(ctx, stmt, params...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var cats []models.Category

	for rows.Next() {
		var c models.Category

		err := rows.Scan(&c.ID, &c.Name)
		if err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}

	return cats, err
}

func GetEventManagers(ctx context.Context, db DBTX, eventId int64) ([]models.Manager, error) {
	query := qb.Select(
		"users.username",
		"users.firstname",
		"users.lastname",
		"users.profile_image",
		"event_managers.user_id",
		"event_managers.role_id",
		"event_roles.name",
	).
		From("users").
		InnerJoin("event_managers on event_managers.user_id = users.user_id").
		InnerJoin("event_roles on event_roles.id = event_managers.role_id").
		Where(sq.Eq{"event_managers.event_id": eventId}).
		Where(sq.Eq{"event_managers.deleted_at": nil})

	stmt, params, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(ctx, stmt, params...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var mans []models.Manager

	for rows.Next() {
		var m models.Manager

		err := rows.Scan(&m.User.Username, &m.User.Firstname, &m.User.Lastname, &m.User.ProfileImage, &m.User.UserID, &m.Role.ID, &m.Role.Name)
		if err != nil {
			return nil, err
		}
		if m.User.ProfileImage != nil {
			*m.User.ProfileImage = pkg.CDNBaseUrl + "/get/" + *m.User.ProfileImage
		}

		m.Role.Permissions, err = GetRolePermissions(ctx, db, m.Role.ID)
		if err != nil {
			return nil, err
		}
		m.Role.EventID = eventId
		m.EventId = eventId
		mans = append(mans, m)
	}

	return mans, nil
}

func RemoveEventCategories(ctx context.Context, db DBTX, eventId int64) error {
	query := qb.Delete("event_categories").Where(sq.Eq{"event_id": eventId})

	stmt, params, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, stmt, params...)
	return err
}

func RemoveManagers(ctx context.Context, db DBTX, eventId int64, managerIds ...int64) error {
	if len(managerIds) == 0 {
		return nil
	}
	query := qb.Update("event_managers").
		Set("deleted_at", time.Now()).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"event_id": eventId}).
		Where(sq.Eq{"id": managerIds})

	stmt, params, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, stmt, params...)
	return err
}

func RemoveLocations(ctx context.Context, db DBTX, eventId int64, locationIds ...int64) error {
	if len(locationIds) == 0 {
		return nil
	}

	query := qb.Update("event_locations").
		Set("deleted_at", time.Now()).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"event_id": eventId}).
		Where(sq.Eq{"id": locationIds})

	stmt, params, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, stmt, params...)
	return err
}

func RemoveImages(ctx context.Context, db DBTX, eventId int64, imgIds ...int64) error {
	if len(imgIds) == 0 {
		return nil
	}

	query := qb.Delete("event_images").
		Where(sq.Eq{"event_id": eventId}).
		Where(sq.Eq{"id": imgIds})

	stmt, params, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, stmt, params...)
	return err
}

func UpdateLocation(ctx context.Context, db DBTX, eventId, locationId int64, location models.Location) error {
	m := map[string]interface{}{}

	if location.Address != nil {
		m["address"] = *location.Address
	}

	if location.Longitude != nil {
		m["longitude"] = *location.Longitude
	}

	if location.Longitude != nil {
		m["latitude"] = *location.Latitude
	}

	if location.Seats != nil {
		m["seats"] = *location.Seats
	}

	if location.StartsAt != nil {
		m["starts_at"] = *location.StartsAt
	}

	if location.EndsAt != nil {
		m["ends_at"] = *location.EndsAt
	}

	query := qb.Update("event").SetMap(m).
		Where(sq.Eq{"id": location}).
		Where(sq.Eq{"event_id": eventId}).
		Where(sq.Eq{"deleted_at": nil})

	stmt, params, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, stmt, params...)
	return err

}

func UpdateMainEvent(ctx context.Context, db DBTX, event models.Event) error {
	m := map[string]interface{}{}
	if event.Title != nil {
		m["title"] = *event.Title
	}
	if event.Description != nil {
		m["description"] = *event.Description
	}
	if event.MaxAge != nil {
		m["age_max"] = *event.MaxAge
	}
	if event.MinAge != nil {
		m["age_min"] = *event.MinAge
	}

	query := qb.Update("event").SetMap(m).
		Where(sq.Eq{"id": event.ID}).
		Where(sq.Eq{"deleted_at": nil})

	stmt, params, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, stmt, params...)
	return err
}

func UpdateManager(ctx context.Context, db DBTX, eventId, managerId int, role models.Role) error {

	query := qb.Update("event_managers").
		Set("role_id", role.ID).
		Where(sq.Eq{"eventId": eventId}).
		Where(sq.Eq{"id": managerId})

	stmt, params, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, stmt, params...)
	return err

}
