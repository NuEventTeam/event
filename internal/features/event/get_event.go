package event

import (
	"context"
	"fmt"
	"github.com/NuEventTeam/events/internal/models"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/gofiber/fiber/v2"
)

func (e Event) GetEventByIDHandler() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		eventID, err := ctx.ParamsInt("eventId")
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
