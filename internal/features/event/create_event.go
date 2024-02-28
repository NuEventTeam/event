package event

import (
	"context"
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

	tx, err := e.db.BeginTx(ctx.Context())
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx.Context())

	eventId, err := database.CreateEvent(ctx.Context(), tx, event)
	if err != nil {
		return err
	}
	err = database.AddEventCategories(ctx.Context(), tx, eventId, request.Categories...)
	if err != nil {
		return err
	}

	var uploadContent []cdn.Content
	images := make([]models.Image, len(form.File["images"]))
	for i, f := range form.File["images"] {
		file, err := f.Open()
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "cannot open file", err)
		}
		filename := ulid.Make().String() + path.Ext(f.Filename)
		uploadContent = append(uploadContent, cdn.Content{
			FieldName: "files",
			Filename:  filename,
			Payload:   file,
			Size:      f.Size,
		})
		event.Images[i].Url = fmt.Sprint(pkg.EventNamespace, "/", eventId, "/", filename)

	}
	err = e.cdn.Upload(fmt.Sprint(pkg.EventNamespace, "/", eventId), event.ImageContent...)
	if err != nil {
		return err
	}

	err = database.AddEventImage(ctx.Context(), tx, eventId, images...)
	if err != nil {
		return err
	}
	err = database.AddEventLocations(ctx.Context(), tx, eventId, event.Locations...)
	if err != nil {
		return err
	}


User: models.User{UserID: userId},
		m.Role.EventID = eventId
		roleId, err := database.CreateRole(ctx.Context(), tx, models.Role{
			Name:        pkg.AuthorTitle,
			Permissions: []int64{pkg.PermissionRead, pkg.PermissionVerify, pkg.PermissionUpdate}})
		if err != nil {
			return err
		}

		m.Role.ID = roleId

		err = database.AddRolePermissions(ctx.Context(), tx, m.Role)
		if err != nil {
			return err
		}

		err = database.AddEventManager(ctx.Context(), tx, eventId, m)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx.Context()); err != nil {
		return err
	}
	eventID, err := e.CreateEvent(ctx.Context(), event)
	if err != nil {
		return pkg.Error(ctx, fiber.StatusInternalServerError, err.Error(), err)
	}

	return pkg.Success(ctx, fiber.Map{"event_id": eventID})
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
	err = database.AddEventCategories(ctx, tx, eventId, event.CategoryIds...)
	if err != nil {
		return 0, err
	}

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
