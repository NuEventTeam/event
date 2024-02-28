package event

import (
	"errors"
	"github.com/NuEventTeam/events/internal/services/cdn"
	"github.com/NuEventTeam/events/internal/storage/cache"
	"github.com/NuEventTeam/events/internal/storage/database"
)

var (
	ErrNoPermission = errors.New("user has no permission")
)

type EventSvc struct {
	db    *database.Database
	cache *cache.Cache
	cdn   *cdn.CdnSvc
}

func NewEventSvc(db *database.Database, cache *cache.Cache, cdn *cdn.CdnSvc) *EventSvc {
	return &EventSvc{
		db:    db,
		cache: cache,
		cdn:   cdn,
	}
}
