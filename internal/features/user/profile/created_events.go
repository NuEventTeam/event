package user_profile

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/NuEventTeam/events/pkg/types"
	"time"
)

func GetOwnEvents(ctx context.Context, db database.DBTX, userId, lastEventId int64) (map[int64]FollowedEvent, error) {
	query := qb.Select(` 
					events.id,
					events.title, 
					events.like_count,
					events.price,
					event_locations.address,
					event_locations.starts_at,
					event_locations.ends_at,
					event_locations.attendees_count`).
		From("events").
		InnerJoin("event_locations on event_locations.event_id = events.id").
		InnerJoin("event_managers on event_managers.event_id = event_locations.event_id").
		InnerJoin("event_roles on event_locations.event_id = event_roles.event_id").
		InnerJoin("event_role_permissions on event_roles.role_id = event_managers.role_id").
		Where(sq.Eq{"event_role_permissions.permission_id": pkg.PermissionUpdate}).
		Where(sq.Eq{"event_managers.user_id": userId})

	if lastEventId != 0 {
		query = query.Where(sq.Lt{"events.id": lastEventId})
	}
	query = query.OrderBy("event_locations.starts_at desc")

	stmt, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	followedEventsMap := map[int64]FollowedEvent{}
	followedEventsSlice := []int64{}
	for rows.Next() {
		var (
			f FollowedEvent
			s time.Time
			e time.Time
		)

		err := rows.Scan(&f.ID, &f.Title, &f.LikesCount, &f.Price, &f.Address, &s, &e, &f.AttendeesCount)
		if err != nil {
			return nil, err
		}

		f.StartsAt = f.StartsAt.FromTime(&s)
		f.EndsAt = f.EndsAt.FromTime(&e)
		f.Date = types.Date(s)
		if f.Price != nil {
			*f.Price = *f.Price / 100
		}
		followedEventsMap[f.ID] = f
		followedEventsSlice = append(followedEventsSlice, f.ID)

	}
	err = getEventImages(ctx, db, followedEventsSlice, followedEventsMap)
	if err != nil {
		return nil, err
	}
	return followedEventsMap, nil

}
