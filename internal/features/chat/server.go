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
		WriteTimeout: 20 * time.Second,
		ReadTimeout:  20 * time.Second,
	}

	return srv.ListenAndServe()
}
