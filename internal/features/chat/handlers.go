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
	ServeWs(ChatManager, w, r.WithContext(ctx))
}
