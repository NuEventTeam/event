package search

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/gofiber/fiber/v2"
)

var qb = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func SearchUser(db *database.Database) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		username := ctx.Query("username", "")
		lastId := ctx.QueryInt("lastId", 0)
		usersMap, userIds, err := getUserByUsername(ctx.Context(), db.GetDb(), username, int64(lastId))
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, "oops something went wrong", err)
		}
		categories, err := getCategories(ctx.Context(), db.GetDb(), userIds)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, "oops something went wrong", err)
		}
		var users []User
		for userId, categories := range categories {
			val, _ := usersMap[userId]

			val.Categories = categories

			users = append(users, val)
		}

		return pkg.Success(ctx, fiber.Map{"users": users})
	}
}

type Categories struct {
	ID   int64  `json:"categoryId"`
	Name string `json:"categoryName"`
}

type User struct {
	ID           int64        `json:"userId"`
	Username     string       `json:"username"`
	ProfileImage *string      `json:"profileImage"`
	Categories   []Categories `json:"categories,omitempty"`
}

func getUserByUsername(ctx context.Context, db database.DBTX, username string, lastId int64) (map[int64]User, []int64, error) {
	query := `select users.id, users.username, users.profile_image 
				from users
				where username LIKE $2 and id  > $3
				order by users.followers_count 
				limit 15`

	rows, err := db.Query(ctx, query, "%"+username+"%", lastId)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var (
		userIds []int64
		users   map[int64]User
	)

	for rows.Next() {
		var u User
		err := rows.Scan(&u.ID, &u.Username, &u.ProfileImage)
		if err != nil {
			return nil, nil, err
		}
		if u.ProfileImage != nil {
			*u.ProfileImage = pkg.CDNBaseUrl + "/get/" + *u.ProfileImage
		}
		userIds = append(userIds, u.ID)

		users[u.ID] = u
	}
	return users, userIds, nil
}

func getCategories(ctx context.Context, db database.DBTX, userIds []int64) (map[int64][]Categories, error) {
	query := qb.Select("categories.id, categories.name,user_preferences.user_id").
		From("categories").
		InnerJoin("user_preferences on user_preferences.category_id = categories.id").
		Where(sq.Eq{"user_preferences.user_id": userIds})

	stmt, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		categories map[int64][]Categories
	)

	for rows.Next() {
		var (
			c      Categories
			userId int64
		)

		err := rows.Scan(&c.ID, &c.Name, &userId)
		if err != nil {
			return nil, err
		}

		if _, ok := categories[userId]; !ok {
			categories[userId] = []Categories{}
		}
		categories[userId] = append(categories[userId], c)

	}
	return categories, nil

}
