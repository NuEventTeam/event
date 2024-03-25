package event

import (
	"context"
	"fmt"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/gofiber/fiber/v2"
	"strconv"
)

func (e *Event) FollowEvent() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		userId := ctx.Locals("userId").(int64)

		eventId, err := strconv.ParseInt(ctx.Params("eventId"), 10, 64)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "invalid event id", err)
		}

		err = e.CheckEventStatus(ctx.Context(), eventId)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
		}

		err = e.addFollower(ctx.Context(), eventId, userId)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "something went wrong", err)
		}

		return pkg.Success(ctx, nil)
	}
}

func (e *Event) addFollower(ctx context.Context, eventId, followerId int64) error {

	tx, err := e.db.BeginTx(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	err = database.AddEventFollower(ctx, tx, eventId, followerId)
	if err != nil {
		return err
	}

	err = database.UpdateEventFollowerCount(ctx, tx, eventId, 1)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (e *Event) CheckEventStatus(ctx context.Context, eventId int64) error {
	event, err := database.GetEventByID(ctx, e.db.GetDb(), eventId)
	if err != nil {
		return err
	}

	if event == nil {
		return fmt.Errorf("event not exists")
	}
	location, err := database.GetEventLocations(ctx, e.db.GetDb(), eventId)
	if err != nil {
		return err
	}

	if location[0].Seats != nil {
		if *location[0].Seats == *location[0].AttendeesCount {
			return fmt.Errorf("event is full")
		}
	}

	if *event.Status != pkg.EventStatusCreated {
		return fmt.Errorf("event cannot be foullowed, status:%d", *event.Status)
	}

	return nil

}
