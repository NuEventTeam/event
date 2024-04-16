package user_follow

import (
	"context"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/gofiber/fiber/v2"
)

func ListFollowed(db *database.Database) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		userId := ctx.Locals("userId").(int64)
		list, err := GetFollowed(ctx.Context(), db.GetDb(), userId)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
		}

		return pkg.Success(ctx, fiber.Map{"followed": list})
	}
}

func GetFollowed(ctx context.Context, db database.DBTX, userId int64) ([]Follower, error) {
	query := `select users.id, users.username, users.profile_image 
				from users
				inner join user_followers on users.id = user_followers.user_id
				where user_followers.follower_id = $1`

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
