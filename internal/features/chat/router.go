package chat

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func getRouter() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/wss/{eventId}", Authorize(joinChatHandler)).Methods("GET")
	r.HandleFunc("/test", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprintf(writer, "OK")
		return
	}).Methods("GET")

	return r
}
