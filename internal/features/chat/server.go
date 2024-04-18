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
		WriteTimeout: 60 * time.Second,
		ReadTimeout:  60 * time.Second,
	}

	return srv.ListenAndServeTLS("./certs/server.crt", "./certs/server.key")
}
