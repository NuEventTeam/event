package event_service

import (
	"context"
	"errors"
	"fmt"
	"github.com/NuEventTeam/events/internal/models"
	"github.com/NuEventTeam/events/internal/storage/cache"
	"github.com/NuEventTeam/events/internal/storage/database"
	"log"
)

var (
	ErrNoPermision = errors.New("user has no permission")
)

type EventSvc struct {
	db    *database.Database
	cache *cache.Cache
}

func NewEventSvc(db *database.Database, cache *cache.Cache) *EventSvc {
	return &EventSvc{
		db:    db,
		cache: cache,
	}
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

func (e *EventSvc) GetEventByID(ctx context.Context, eventId int64) (*models.Event, error) {
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

func (e *EventSvc) GetCategoriesByID(ctx context.Context, ids []int64) ([]models.Category, error) {
	categories, err := database.GetCategories(ctx, e.db.GetDb(), database.GetCategoriesParams{IDs: ids})
	return categories, err
}

func (e *EventSvc) UpdateEvent(ctx context.Context, event models.Event) error {
	tx, err := e.db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = database.UpdateMainEvent(ctx, tx, event)
	if err != nil {
		return err
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
	//update location
	if len(event.Locations) > 0 {
		err := database.UpdateLocation(ctx, tx, event.ID, event.Locations[0].ID, event.Locations[0])
		if err != nil {
			return err
		}
	}

	if len(event.Images) > 0 {
		err := database.AddEventImage(ctx, tx, event.ID, event.Images...)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (e *EventSvc) RemoveImage(ctx context.Context, eventID int64, imgIds ...int64) ([]string, error) {

	images, err := database.GetEventImages(ctx, e.db.GetDb(), eventID, imgIds...)
	if err != nil {
		return nil, err
	}

	err = database.RemoveImages(ctx, e.db.GetDb(), eventID, imgIds...)
	if err != nil {
		return nil, err
	}

	var keys []string

	for _, i := range images {
		keys = append(keys, i.Url)
	}

	return keys, nil
}

func (e *EventSvc) CheckPermission(ctx context.Context, eventId, userId int64, permissionIds ...int64) error {
	ok, err := database.CheckPermission(ctx, e.db.GetDb(), eventId, userId, permissionIds...)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNoPermision
	}
	return nil
}

func (e *EventSvc) AddFollower(ctx context.Context, eventId, followerId int64) error {
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

func (e *EventSvc) RemoveFollower(ctx context.Context, eventId, followerId int64) error {
	tx, err := e.db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	err = database.RemoveEventFollower(ctx, tx, eventId, followerId)
	if err != nil {
		return err
	}
	err = database.UpdateEventFollowerCount(ctx, tx, eventId, -1)
	if err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

//TODO finish later
//func (e *EventSvc) BanFollower(ctx context.Context, eventId, followerId int64) error {
//	tx, err := e.db.BeginTx(ctx)
//	if err != nil {
//		return err
//	}
//	defer tx.Rollback(ctx)
//	err = database.BanEventFollower(ctx, tx, eventId, followerId)
//	if err != nil {
//		return err
//	}
//	//Todo check if it was user follower ad decrease follower count
//
//	if err := tx.Commit(ctx); err != nil {
//		return err
//	}
//	return nil
//}
