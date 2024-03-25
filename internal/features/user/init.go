package user

import (
	"errors"
	"github.com/Masterminds/squirrel"
	"github.com/NuEventTeam/events/internal/features/assets"
	"github.com/NuEventTeam/events/internal/storage/database"
)

var qb = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

var (
	ErrNoPermission = errors.New("user has no permission")
)

type User struct {
	db     *database.Database
	assets *assets.Assets
}

func NewEventSvc(db *database.Database, assets *assets.Assets) *User {
	return &User{
		db:     db,
		assets: assets,
	}
}
