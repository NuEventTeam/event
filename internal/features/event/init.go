package event

import (
	"errors"
	"github.com/Masterminds/squirrel"
	"github.com/NuEventTeam/events/internal/services/cdn"
	"github.com/NuEventTeam/events/internal/storage/cache"
	"github.com/NuEventTeam/events/internal/storage/database"
)

var qb = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

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
