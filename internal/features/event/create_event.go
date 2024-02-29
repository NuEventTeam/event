package event

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/NuEventTeam/events/internal/models"
	"github.com/NuEventTeam/events/internal/services/cdn"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/oklog/ulid/v2"
	"log"
	"path"
)

func RegisterRouter()

type CreateEventRequest struct {
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Price       *float32      `json:"price"`
	Seats       *int64        `json:"seats"`
	MaxAge      *int64        `json:"max_age"`
	MinAge      *int64        `json:"min_age"`
	Address     string        `json:"address"`
	Longitude   float64       `json:"lg"`
	Latitude    float64       `json:"lt"`
	Date        string        `json:"date"`
	StartsAt    *pkg.DateTime `json:"starts_at"`
	EndsAt      *pkg.DateTime `json:"end_at"`
	Categories  []int64       `json:"categories"`
}

func (e *EventSvc) createEventHttp(ctx *fiber.Ctx) error {
	form, err := ctx.MultipartForm()
	if err != nil {
		return err
	}

	userId := ctx.Locals("userId").(int64)

	if len(form.Value["payload"]) == 0 {
		return pkg.Error(ctx, fiber.StatusBadRequest, "request does not contain payload field", fmt.Errorf("payload missing"))
	}
	payload := form.Value["payload"][0]

	var request CreateEventRequest

	err = sonic.ConfigFastest.Unmarshal([]byte(payload), &request)
	if err != nil {
		return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
	}

	uploadContent := make([]cdn.Content, len(form.File["images"]))
	for i, f := range form.File["images"] {
		file, err := f.Open()
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "cannot open file", err)
		}
		filename := ulid.Make().String() + path.Ext(f.Filename)
		uploadContent[i] = cdn.Content{
			FieldName: "files",
			Filename:  filename,
			Payload:   file,
			Size:      f.Size,
		}
	}

	err = func(ctx context.Context) error {
		tx, err := e.db.BeginTx(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)
		runner := qb.RunWith(&sql.DB{})

		err = runner.Insert("events").
			Columns("title", "description", "age_min", "age_max").
			Values(request.Title, request.Description, request.MinAge, request.MaxAge).
			Suffix("RETURNING id").QueryRow().Scan()
		if err != nil {
			return err
		}
		var eventId int64
		err = tx.QueryRow(ctx, stmt, args...).Scan(&eventId)
		if err != nil {
			return err
		}

		query := qb.Insert("event_categories").
			Columns("event_id", "category_id")

		for _, id := range request.Categories {
			query = query.Values(eventId, id)
		}

		stmt, args, err = query.ToSql()
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx, stmt, args...)
		if err != nil {
			return err
		}

		err = e.cdn.Upload(fmt.Sprint(pkg.EventNamespace, "/", eventId), uploadContent...)
		if err != nil {
			return err
		}

		query = qb.Insert("event_images").
			Columns("event_id", "url")
		for _, upl := range uploadContent {
			query = query.Values(eventId, pkg.EventNamespace+"/"+upl.Filename)
		}

		stmt, args, err = query.ToSql()
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx, stmt, args...)
		if err != nil {
			return err
		}

		if err := tx.Commit(ctx); err != nil {
			return err
		}
		return nil

	}(ctx.Context())
	return nil
}

func (e *EventSvc) CreateEvent(ctx context.Context, event models.Event) (int64, error) {
	tx, err := e.db.BeginTx(ctx)
	if err != nil {
		return 0, err
	}

	defer tx.Rollback(ctx)

	eventId, err := database.CreateEvent(ctx, tx, event)
	if err != nil {
		return 0, err
	}
	log.Println(eventId)

	err = e.cdn.Upload(fmt.Sprint(pkg.EventNamespace, "/", eventId), event.ImageContent...)
	if err != nil {
		return 0, err
	}

	for i, c := range event.ImageContent {
		event.Images[i].Url = fmt.Sprint(pkg.EventNamespace, "/", eventId, "/", c.Filename)
	}

	err = database.AddEventImage(ctx, tx, eventId, event.Images...)
	if err != nil {
		return 0, err
	}
	err = database.AddEventLocations(ctx, tx, eventId, event.Locations...)
	if err != nil {
		return 0, err
	}

	for _, m := range event.Managers {
		log.Println(m.User.UserID)
		m.Role.EventID = eventId
		roleId, err := database.CreateRole(ctx, tx, m.Role)
		if err != nil {
			return 0, err
		}

		m.Role.ID = roleId

		err = database.AddRolePermissions(ctx, tx, m.Role)
		if err != nil {
			return 0, err
		}

		err = database.AddEventManager(ctx, tx, eventId, m)
		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return eventId, nil
}
