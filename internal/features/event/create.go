package event

import (
	"context"
	"fmt"
	"github.com/NuEventTeam/events/internal/features/assets"
	"github.com/NuEventTeam/events/internal/models"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/NuEventTeam/events/pkg/types"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/oklog/ulid/v2"
	"log"
	"path"
	"sync"
	"time"
)

type CreateEventRequest struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Price       *float32       `json:"price"`
	Seats       *int64         `json:"seats"`
	MaxAge      *int64         `json:"max_age"`
	MinAge      *int64         `json:"min_age"`
	Address     string         `json:"address"`
	Longitude   float64        `json:"lg"`
	Latitude    float64        `json:"lt"`
	Date        types.DateTime `json:"date"`
	StartsAt    types.DateTime `json:"starts_at"`
	EndsAt      types.DateTime `json:"end_at"`
	Categories  []int64        `json:"categories"`
}

func (c *CreateEventRequest) FromPayload(payload []byte) error {
	err := sonic.ConfigFastest.Unmarshal([]byte(payload), &c)
	if err != nil {
		return err
	}
	return nil
}
func (c *CreateEventRequest) Validate() []string {
	m := []string{}

	if c.Date.Before(time.Now()) {
		m = append(m, "date cannot be set before current date")
	}

	if c.StartsAt.Before(time.Now().Add(time.Hour * 2)) {
		m = append(m, "event start time mast be at least 2 hour before creations")
	}

	if c.EndsAt.Before(time.Time(c.StartsAt)) {
		m = append(m, "improper ending time")
	}
	if len(m) > 0 {
		return m
	}
	return nil
}

func (e *Event) CreateEventHandler() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		form, err := ctx.MultipartForm()
		if err != nil {
			return err
		}

		userId := ctx.Locals("userId").(int64)

		msg, ok := requireFormFiled(form, "payload")
		if !ok {
			return pkg.Error(ctx, fiber.StatusBadRequest, msg)
		}

		payload := form.Value["payload"][0]

		var request CreateEventRequest

		err = request.FromPayload([]byte(payload))
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, err.Error())
		}

		violations := request.Validate()
		if violations != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, violations)
		}

		images := make([]*assets.Image, len(form.File["images"]))
		wg := sync.WaitGroup{}

		for i, f := range form.File["images"] {
			wg.Add(1)
			f := f
			go func(index int) {
				defer wg.Done()
				file, err := f.Open()
				if err != nil {
					log.Println("cannot open the image file")
					return
				}

				filename := ulid.Make().String() + path.Ext(f.Filename)

				img, err := assets.NewImage(filename, file, assets.WithWidthAndHeight(500, 500))

				if err != nil {
					log.Println("while uploading image")
				} else {
					images[index] = img
				}
			}(i)
		}

		wg.Wait()

		event := models.Event{
			Title:       &request.Title,
			Description: &request.Description,
			MaxAge:      request.MaxAge,
			MinAge:      request.MinAge,
			Images:      images,
			CategoryIds: request.Categories,
			Locations: []models.Location{{
				Address:   &request.Address,
				Longitude: &request.Longitude,
				Latitude:  &request.Latitude,
				Seats:     request.Seats,
				StartsAt:  &request.StartsAt,
				EndsAt:    &request.EndsAt,
			}},
			Managers: []models.Manager{{
				User: models.User{UserID: userId},
				Role: models.Role{
					Name:        pkg.AuthorTitle,
					Permissions: []int64{pkg.PermissionRead, pkg.PermissionVerify, pkg.PermissionUpdate}},
			}},
			Attendees: nil,
		}
		eventID, err := e.createEvent(ctx.Context(), event)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, err.Error(), err)
		}

		return pkg.Success(ctx, fiber.Map{"event_id": eventID})
	}
}

func (e Event) createEvent(ctx context.Context, event models.Event) (int64, error) {
	tx, err := e.db.BeginTx(ctx)
	if err != nil {
		return 0, err
	}

	defer tx.Rollback(ctx)

	eventId, err := database.CreateEvent(ctx, tx, event)
	if err != nil {
		return 0, err
	}

	err = database.AddEventCategories(ctx, tx, eventId, event.CategoryIds...)
	if err != nil {
		return 0, err
	}

	for i, c := range event.Images {
		event.Images[i].Filename = fmt.Sprint(pkg.EventNamespace, "/", eventId, "/", c.Filename)
	}

	err = database.AddEventImage(ctx, tx, eventId, event.Images...)
	if err != nil {
		return 0, err
	}

	e.assets.Upload(ctx, event.Images...)

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
