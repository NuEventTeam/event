package handlers

import (
	"github.com/NuEventTeam/events/internal/features/assets"
	"github.com/NuEventTeam/events/internal/features/auth"
	"github.com/NuEventTeam/events/internal/features/event"
	"github.com/NuEventTeam/events/internal/features/user"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/internal/storage/keydb"
)

type Handler struct {
	EventSvc  *event.Event
	Cache     *keydb.Cache
	DB        *database.Database
	UserSvc   *user.User
	Assets    *assets.Assets
	Auth      *auth.Auth
	JwtSecret string
}

func New(event *event.Event, cache *keydb.Cache, user *user.User, assets *assets.Assets, auth *auth.Auth, jwt string, db *database.Database) *Handler {
	return &Handler{
		EventSvc:  event,
		Cache:     cache,
		UserSvc:   user,
		Assets:    assets,
		Auth:      auth,
		JwtSecret: jwt,
		DB:        db,
	}
}
