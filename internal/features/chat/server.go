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
		WriteTimeout: 1 * time.Second,
		ReadTimeout:  1 * time.Second,
	}

	//return srv.ListenAndServeTLS("/etc/ssl/certs/apache-selfsigned.crt", "/etc/ssl/private/apache-selfsigned.key")
	return srv.ListenAndServe()
}
