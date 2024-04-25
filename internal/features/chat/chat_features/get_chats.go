package chat_features

import (
	"context"
	"fmt"
	"github.com/Masterminds/squirrel"
	user_profile "github.com/NuEventTeam/events/internal/features/user/profile"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/gofiber/fiber/v2"
	"log"
	"time"
)

var qb = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

type Chat struct {
	EventID     int64     `json:"eventId"`
	Title       string    `json:"title"`
	Images      []string  `json:"images"`
	LastMessage *Messages `json:"lastMessage"`
}

func GetChats(db *database.Database) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		userId := ctx.Locals("userId").(int64)
		lastId := ctx.QueryInt("lastId", 0)
		followed, err := user_profile.GetFollowedEvents(ctx.Context(), db.GetDb(), userId, int64(lastId))
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, "oops error", err)
		}

		lastMessages, err := getLastMessages(ctx.Context(), db.GetDb(), userId)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, "oops errors", err)
		}

		chats := []Chat{}
		log.Println(followed)
		log.Println(lastMessages)
		for key, val := range followed {
			lastMessage := lastMessages[key]
			chats = append(chats, Chat{
				EventID:     val.ID,
				Images:      val.Images,
				Title:       val.Title,
				LastMessage: &lastMessage,
			})
		}

		return pkg.Success(ctx, fiber.Map{"chatInfo": chats})
	}
}

type Messages struct {
	ID           int64     `json:"id"`
	EventId      int64     `json:"eventId"`
	UserId       int64     `json:"userId"`
	Username     string    `json:"username"`
	ProfileImage *string   `json:"profileImage"`
	Message      string    `json:"message"`
	CreatedAt    time.Time `json:"createdAt"`
	IsMy         bool      `json:"isMy"`
}

func getLastMessages(ctx context.Context, db database.DBTX, userId int64) (map[int64]Messages, error) {
	query := `
select  chat_members.event_id from chat_members
		inner join chat_messages on chat_messages.event_id = event_followers.event_id
		where chat_members.user_id = $1 group by event_followers.event_id
`

	eventIds := []int64{}

	rows, err := db.Query(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int64
		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		eventIds = append(eventIds, id)
	}
	rows.Close()

	q := qb.Select(`chat_messages.id, chat_messages.event_id,user_id,username,profile_image,chat_messages.created_at, messages`).
		From("chat_messages").
		InnerJoin("users on users.id = chat_messages.user_id").
		Where(squirrel.Eq{"chat_messages.event_id": eventIds}).OrderBy("chat_messages desc")
	stmt, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}
	log.Println(stmt)

	messages := map[int64]Messages{}
	rows, err = db.Query(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var m Messages
		err := rows.Scan(&m.ID, &m.EventId, &m.UserId, &m.Username, &m.ProfileImage, &m.CreatedAt, &m.Message)
		if err != nil {
			return nil, err
		}
		if m.ProfileImage != nil {
			profileImgUrl := fmt.Sprint(pkg.CDNBaseUrl, *m.ProfileImage)
			m.ProfileImage = &profileImgUrl
		}
		if m.UserId == userId {
			m.IsMy = true
		}
		if _, ok := messages[m.EventId]; !ok {
			messages[m.EventId] = m
		}

	}

	return messages, nil
}
