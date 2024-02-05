package database

import (
	"context"
	"errors"
	sq "github.com/Masterminds/squirrel"
	"github.com/NuEventTeam/events/internal/models"
	"github.com/jackc/pgx/v5"
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

	return 0, err
}

func AddEventManager(ctx context.Context, db DBTX, eventID int64, manager ...models.Manager) error {

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

func AddEventCategories(ctx context.Context, db DBTX, eventID int64, category ...models.Category) error {

	query := qb.Insert("event_categories").
		Columns("event_id", "category_id")

	for i := range category {
		query = query.Values(eventID, category[i].ID)
	}

	stmt, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, stmt, args...)
	return err
}

func AddEventLocations(ctx context.Context, db DBTX, eventID int64, locations ...models.Location) error {
	query := qb.Insert("event_locations").
		Columns("event_id", "address", "longitude", "latitude", "seats", "starts_at", "ends_at")

	for _, l := range locations {
		query = query.Values(eventID, l.Address, l.Longitude, l.Latitude, l.Seats, l.StartsAt, l.EndsAt)
	}

	stmt, params, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, stmt, params...)

	return err

}

func AddEventImage(ctx context.Context, db DBTX, eventID int64, images ...models.Image) error {
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
	query := qb.Select("events.id", "title", "description", "age_min", "age_max", "created_at").
		From("events").Where(sq.Eq{"id": eventID})

	stmt, params, err := query.ToSql()
	if err != nil {
		return nil, err
	}

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

func GetEventImages(ctx context.Context, db DBTX, eventId int64) ([]models.Image, error) {
	query := qb.Select("id", "event_id", "url").
		From("images").
		Where(sq.Eq{"deleted_at": nil}).
		Where(sq.Eq{"event_id": eventId})

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
	query := qb.Select("users.username", "event_manages.user_id, event_managers.title, event_managers.role_id, event_roles.name").
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

		err := rows.Scan(&m.User.Username, &m.User.UserID, &m.Title, &m.Role.ID, m.Role.Name)
		if err != nil {
			return nil, err
		}

		mans = append(mans, m)
	}

	return mans, nil
}
