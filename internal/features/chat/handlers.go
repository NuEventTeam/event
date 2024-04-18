package chat

import (
	"context"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

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

//
//func GetChats(db *database.Database) fiber.Handler {
//	return func(ctx *fiber.Ctx) error {
//		userId := ctx.Locals("userId").(int64)
//		lastId := ctx.Q
//		followed, err := user_profile.GetFollowedEvents(ctx.Context(), db.GetDb(), userId)
//		if err != nil {
//			return pkg.Error(ctx, fiber.StatusInternalServerError, "oops error", err)
//		}
//
//	}
//}
//
//func getChats(ctx context.Context, db database.DBTX, userId int64) {
//
//}
