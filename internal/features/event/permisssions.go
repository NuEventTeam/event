package event

import (
	"context"
	"github.com/NuEventTeam/events/internal/storage/database"
)

func (e *Event) CheckPermission(ctx context.Context, eventId, userId int64, permissionIds ...int64) error {
	ok, err := database.CheckPermission(ctx, e.db.GetDb(), eventId, userId, permissionIds...)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNoPermission
	}
	return nil
}
