package chat

import (
	"fmt"
	"net/http"
	"time"
)

func RunChatServer(port int) error {
	ChatManager = NewManager()
	go ChatManager.Run()

	srv := &http.Server{
		Handler:      getRouter(),
		Addr:         fmt.Sprintf(":%d", port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	return srv.ListenAndServe()
}
