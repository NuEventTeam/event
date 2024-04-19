package chat

import (
	"context"
	"fmt"
	"github.com/Masterminds/squirrel"
	user_profile "github.com/NuEventTeam/events/internal/features/user/profile"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/gofiber/fiber/v2"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"time"
)

var qb = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

func joinChatHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("here")

	eventID, err := strconv.ParseInt(mux.Vars(r)["eventId"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("something went wrogn"))
		log.Println(err)
		return
	}
	ctx := context.WithValue(r.Context(), "eventId", eventID)
	//ctx = context.WithValue(ctx, "userId", rand.Int63())

	ServeWs(ChatManager, w, r.WithContext(ctx))
}

type Chat struct {
	EventID     int64    `json:"eventId"`
	Title       string   `json:"title"`
	Images      []string `json:"images"`
	LastMessage Messages `json:"lastMessage"`
}

func GetChats(db *database.Database) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		userId := ctx.Locals("userId").(int64)
		lastId := ctx.QueryInt("lastId")
		followed, err := user_profile.GetFollowedEvents(ctx.Context(), db.GetDb(), userId, int64(lastId))
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, "oops error", err)
		}

		lastMessages, err := getLastMessages(ctx.Context(), db.GetDb(), userId)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, "oops errors", err)
		}

		chats := []Chat{}

		for key, val := range followed {
			lastMessage := lastMessages[key]
			chats = append(chats, Chat{
				EventID:     key,
				Images:      val.Images,
				Title:       val.Title,
				LastMessage: lastMessage,
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
	query := `select  event_id from event_followers inner join chat_messages on chat_messages.event_id = event_followers.event_id
group by  event_id where user_id = $1`

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

	q := qb.Select("id, event_id,user_id, messages, max(chat_messages.created_at), username, profile_images").
		From("chat_messages").InnerJoin("users on users.id = chat_messages.user_id").
		Where(squirrel.Eq{"event_id": eventIds}).OrderBy("created_at desc")

	stmt, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	messages := map[int64]Messages{}
	rows, err = db.Query(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var m Messages
		err := rows.Scan(&m.ID, &m.EventId, &m.UserId, &m.Message, &m.CreatedAt, &m.Username, &m.ProfileImage)
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
		messages[m.EventId] = m

	}

	return messages, nil
}

func GetChatMessages(db *database.Database) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		userId := ctx.Locals("userid").(int64)
		eventId, err := ctx.ParamsInt("eventId")
		lastId := ctx.QueryInt("lastId")
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
		}
		messages, err := FetchChatMessage(ctx.Context(), db.GetDb(), int64(eventId), userId, int64(lastId))
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, err.Error(), err)
		}

		return pkg.Success(ctx, fiber.Map{"messages": messages})
	}
}

func FetchChatMessage(ctx context.Context, db database.DBTX, eventId int64, userId, lastId int64) ([]Messages, error) {
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

		if m.UserId == userId {
			m.IsMy = true
		}
		messages = append(messages, m)
	}

	return nil, err
}
