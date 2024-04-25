package chat_features

import (
	"context"
	"fmt"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"time"
)

func SaveMessage(db database.DBTX, eventId, userId int64, message string) (Messages, error) {

	query := ` insert into chat_messages(event_id,user_id,messages) values($1,$2,$3) returning id,created_at`
	var (
		messageId int64
		createdAt time.Time
	)
	err := db.QueryRow(context.Background(), query, eventId, userId, message).Scan(&messageId, &createdAt)
	if err != nil {
		return Messages{}, err
	}

	query = `select username,profile_image from users where id = $1`

	var (
		username     string
		profileImage *string
	)

	err = db.QueryRow(context.Background(), query, userId).Scan(&username, &profileImage)
	if err != nil {
		return Messages{}, err
	}

	if profileImage != nil {
		profileImgUrl := fmt.Sprint(pkg.CDNBaseUrl, *profileImage)
		profileImage = &profileImgUrl
	}
	return Messages{
		ID:           messageId,
		CreatedAt:    createdAt,
		UserId:       userId,
		Username:     username,
		ProfileImage: profileImage,
		Message:      message,
		EventId:      eventId,
	}, nil
}
