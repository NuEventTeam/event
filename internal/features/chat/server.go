package chat

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func RunChatServer(port int) error {
	log.Println("staring chat", port)

	ChatManager = NewManager()
	go ChatManager.Run()
	srv := &http.Server{
		Handler:      getRouter(),
		Addr:         fmt.Sprintf(":%d", port),
		WriteTimeout: 60 * time.Second,
		ReadTimeout:  60 * time.Second,
	}
	//return srv.ListenAndServe()
	return srv.ListenAndServeTLS("./certs/domain.crt", "./certs/domain.key")
}
