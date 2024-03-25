package chat

import (
	"github.com/gorilla/mux"
)

func getRouter() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/ws/{eventId}", Authorize(joinChatHandler)).Methods("GET")

	return r
}
