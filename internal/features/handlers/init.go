package handlers

import (
	"github.com/NuEventTeam/events/internal/features/event"
	"github.com/NuEventTeam/events/internal/features/user"
)

type Handler struct {
	EventSvc  *event.Event
	UserSvc   *user.User
	JwtSecret string
}

func New() {

}
