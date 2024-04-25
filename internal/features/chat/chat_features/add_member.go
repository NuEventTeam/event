package chat_features

import (
	"context"
	"github.com/NuEventTeam/events/internal/storage/database"
)

func getUsernameByID(ctx context.Context, db database.DBTX, userId int64) (string, error) {
	query := `select username from users where id = $1`
	var username string
	err := db.QueryRow(ctx, query, userId).Scan(&username)
	return username, err
}
func AddChatMember(ctx context.Context, db database.DBTX, eventId, userId, roleId int64) error {
	err := addMember(ctx, db, eventId, userId, roleId)
	if err != nil {
		return err
	}
	//
	//username, err := getUsernameByID(ctx, db, userId)
	//if err != nil {
	//	return err
	//}
	//
	//msg, err := SaveMessage(db, eventId, 0, fmt.Sprintf("%s has joined the chat", username))
	//if err != nil {
	//	return err
	//}
	//
	//members := chat.ChatManager.EventList[eventId]
	//js, err := sonic.ConfigFastest.Marshal(msg)
	//if err != nil {
	//	return err
	//}
	//for _, m := range members {
	//	m := m
	//	go func() {
	//		m.SendMsgChan <- chat.Message{
	//			EventId: eventId,
	//			Payload: js,
	//			From:    0,
	//		}
	//	}()
	//}
	return nil
}

func addMember(ctx context.Context, db database.DBTX, eventId, userId, roleId int64) error {
	query := `insert into chat_members (user_id, event_id, role_id) values($1,$2,$3) `

	args := []interface{}{eventId, userId, roleId}

	_, err := db.Exec(ctx, query, args...)

	return err
}
