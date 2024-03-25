package handlers

import (
	"github.com/NuEventTeam/events/internal/features/assets"
	"github.com/NuEventTeam/events/internal/features/auth"
	"github.com/NuEventTeam/events/internal/features/event"
	"github.com/NuEventTeam/events/internal/features/user"
)

type Handler struct {
	EventSvc  *event.Event
	UserSvc   *user.User
	Assets    *assets.Assets
	Auth      *auth.Auth
	JwtSecret string
}

func New(event *event.Event, user *user.User, assets *assets.Assets, auth *auth.Auth, jwt string) *Handler {
	return &Handler{
		EventSvc:  event,
		UserSvc:   user,
		Assets:    assets,
		Auth:      auth,
		JwtSecret: jwt,
	}
}
