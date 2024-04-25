package chat_features

import (
	"context"
	"fmt"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/gofiber/fiber/v2"
)

func GetChatMessages(db *database.Database) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		userId := ctx.Locals("userId").(int64)
		eventId, err := ctx.ParamsInt("eventId")
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
		}
		lastId := ctx.QueryInt("lastId", 0)

		messages, err := FetchChatMessage(ctx.Context(), db.GetDb(), int64(eventId), int64(lastId))
		for i, msg := range messages {
			if msg.UserId == userId {
				messages[i].IsMy = true
			}
		}
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, err.Error(), err)
		}

		return pkg.Success(ctx, fiber.Map{"messages": messages})
	}
}

func FetchChatMessage(ctx context.Context, db database.DBTX, eventId int64, lastId int64) ([]Messages, error) {
	query := `select chat_messages.id, user_id,username,profile_image,chat_messages.created_at, messages
				from chat_messages inner join users on users.id = chat_messages.user_id
				where chat_messages.event_id = $1 and chat_messages.id > $2
				order by id desc
`

	rows, err := db.Query(ctx, query, eventId, lastId)
	if err != nil {
		return nil, err
	}

	var messages []Messages

	defer rows.Close()
	for rows.Next() {
		var m Messages
		err := rows.Scan(&m.ID, &m.UserId, &m.Username, &m.ProfileImage, &m.CreatedAt, &m.Message)
		if err != nil {
			return nil, err
		}
		if m.ProfileImage != nil {
			profileImgUrl := fmt.Sprint(pkg.CDNBaseUrl, *m.ProfileImage)
			m.ProfileImage = &profileImgUrl
		}

		messages = append(messages, m)
	}

	return messages, nil
}
