package event

import (
	"context"
	"fmt"
	"github.com/NuEventTeam/events/internal/models"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/NuEventTeam/events/pkg/types"
	"github.com/gofiber/fiber/v2"
)

type UpdateEventRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Location    *struct {
		ID        int64           `json:"locationId"`
		Address   *string         `json:"address"`
		Longitude *float64        `json:"longitude"`
		Latitude  *float64        `json:"latitude"`
		StartsAt  *types.DateTime `json:"startsAt"`
		EndsAt    *types.DateTime `json:"endsAt"`
		Seats     *int64          `json:"seats"`
	} `json:"location"`
	Categories     []int64 `json:"categories"`
	AgeMax         *int64  `json:"ageMax"`
	AgeMin         *int64  `json:"ageMin"`
	Status         *int    `json:"status"`
	RemoveImageIds []int64 `json:"removeImages"`
}

func (e Event) UpdateEventHandler() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var request UpdateEventRequest

		if err := ctx.BodyParser(&request); err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "invalid json", err)
		}

		eventId, err := ctx.ParamsInt("eventId")
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "invalid event id", err)
		}

		event := models.Event{
			ID:              int64(eventId),
			Title:           request.Title,
			Description:     request.Description,
			MaxAge:          request.AgeMax,
			MinAge:          request.AgeMin,
			CategoryIds:     request.Categories,
			Status:          request.Status,
			RemoveImagesIds: request.RemoveImageIds,
		}

		if request.Location != nil {
			l := request.Location
			var startsAt, endsAt *types.DateTime
			if l.StartsAt != nil {
				startsAt = l.StartsAt
			}
			if l.EndsAt != nil {
				endsAt = l.EndsAt
			}
			event.Locations = append(event.Locations, models.Location{
				ID:        l.ID,
				EventID:   int64(eventId),
				Address:   l.Address,
				Longitude: l.Longitude,
				Latitude:  l.Latitude,
				StartsAt:  startsAt,
				EndsAt:    endsAt,
				Seats:     l.Seats,
			})
		}

		err = e.UpdateEvent(ctx.Context(), event)
		if err != nil {
			return err
		}
		return pkg.Success(ctx, nil)
	}
}

func (e Event) UpdateEvent(ctx context.Context, event models.Event) error {
	tx, err := e.db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = database.UpdateMainEvent(ctx, tx, event)
	if err != nil {
		return err
	}

	if len(event.RemoveImagesIds) > 0 {
		imgs, err := database.GetEventImages(ctx, tx, event.ID, event.ImageIds...)
		if err != nil {
			return err
		}

		err = database.RemoveImages(ctx, tx, event.ID, event.ImageIds...)
		if err != nil {
			return err
		}

		urls := make([]string, len(imgs))
		for _, i := range imgs {
			urls = append(urls, i.Url)
		}

		e.assets.DeleteFile(ctx, urls...)
	}

	if len(event.CategoryIds) > 0 {
		err := database.RemoveEventCategories(ctx, tx, event.ID)
		if err != nil {
			return err
		}

		err = database.AddEventCategories(ctx, tx, event.ID, event.CategoryIds...)
		if err != nil {
			return err
		}
	}

	if len(event.Locations) > 0 {
		err := database.UpdateLocation(ctx, tx, event.ID, event.Locations[0].ID, event.Locations[0])
		if err != nil {
			return err
		}
	}

	if len(event.Images) > 0 {
		for i, c := range event.Images {
			event.Images[i].Filename = fmt.Sprint(pkg.EventNamespace, "/", event.ID, "/", c.Filename)
		}
		err := database.AddEventImage(ctx, tx, event.ID, event.Images...)
		if err != nil {
			return err
		}
		e.assets.Upload(ctx, event.Images...)
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}
