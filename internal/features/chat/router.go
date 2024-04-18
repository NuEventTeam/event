package chat

import (
	"github.com/gorilla/mux"
)

func getRouter() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/test/{eventId}", Authorize(joinChatHandler)).Methods("GET")
	r.HandleFunc("/test", Authorize(joinChatHandler)). //	func(writer http.ResponseWriter, request *http.Request) {
		//	fmt.Fprintf(writer, "OKK")
		//	return
		//}
		Methods("GET")

	return r
}
