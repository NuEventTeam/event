package event_service

import (
	"context"
	"fmt"
	"github.com/NuEventTeam/events/internal/models"
	"github.com/NuEventTeam/events/internal/storage/cache"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/jackc/pgx/v5"
	"log"
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
	tx, err := e.db.DB.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return 0, err
	}

	defer tx.Rollback(ctx)

	eventId, err := database.CreateEvent(ctx, tx, event)
	if err != nil {
		return 0, err
	}
	log.Println(eventId)
	err = database.AddEventCategories(ctx, tx, eventId, event.Categories...)
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
	event, err := database.GetEventByID(ctx, e.db.DB, eventId)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, fmt.Errorf("event not found")
	}
	log.Println(event)
	categories, err := database.GetEventCategories(ctx, e.db.DB, eventId)
	if err != nil {
		return nil, err
	}
	log.Println(categories)

	locations, err := database.GetEventLocations(ctx, e.db.DB, eventId)
	if err != nil {
		return nil, err
	}
	log.Println(locations)

	images, err := database.GetEventImages(ctx, e.db.DB, eventId)
	if err != nil {
		return nil, err
	}
	log.Println(images)

	managers, err := database.GetEventManagers(ctx, e.db.DB, eventId)
	if err != nil {
		return nil, err
	}
	log.Println(managers)

	event.Categories = categories
	event.Locations = locations
	event.Images = images
	event.Managers = managers
	return event, nil
}

func (e *EventSvc) GetCategoriesByID(ctx context.Context, ids []int64) ([]models.Category, error) {
	categories, err := database.GetCategories(ctx, e.db.DB, database.GetCategoriesParams{IDs: ids})
	return categories, err
}
