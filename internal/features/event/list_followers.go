package event

import (
	"context"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/gofiber/fiber/v2"
)

func (e Event) ListFollowers() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		eventId, err := ctx.ParamsInt("eventId")

		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
		}
		list, err := getFollowers(ctx.Context(), e.db.GetDb(), int64(eventId))
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
		}

		return pkg.Success(ctx, fiber.Map{"followers": list})
	}
}

type Follower struct {
	UserId       int64   `json:"userId"`
	Username     string  `json:"username"`
	ProfileImage *string `json:"profileImage"`
}

func getFollowers(ctx context.Context, db database.DBTX, userId int64) ([]Follower, error) {
	query := `select users.id, users.username, users.profile_image 
				from users
				inner join event_follower on users.id = user_followers.follower_id
				where event_id = $1`

	rows, err := db.Query(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var followers []Follower
	for rows.Next() {
		var f Follower
		err := rows.Scan(&f.UserId, &f.Username, &f.ProfileImage)
		if err != nil {
			return nil, err
		}
		if f.ProfileImage != nil {
			*f.ProfileImage = pkg.CDNBaseUrl + *f.ProfileImage
		}
		followers = append(followers, f)
	}
	return followers, nil
}
