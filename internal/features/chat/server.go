package chat

import (
	"fmt"
	"github.com/NuEventTeam/events/internal/storage/database"
	"log"
	"net/http"
	"time"
)

func RunChatServer(port int, db *database.Database) error {
	log.Println("staring chat", port)
	DB = db
	ChatManager = NewManager()
	go ChatManager.Run()
	srv := &http.Server{
		Handler:      getRouter(),
		Addr:         fmt.Sprintf(":%d", port),
		WriteTimeout: 60 * time.Second,
		ReadTimeout:  60 * time.Second,
	}
	//headersOk := hd.AllowedHeaders([]string{"X-Requested-With"})
	//originsOk := hd.AllowedOrigins([]string{os.Getenv("ORIGIN_ALLOWED")})
	//methodsOk := hd.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	// start server listen
	// with error handling
	//log.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%d", port), "./certs/domain.crt", "./certs/domain.key", hd.CORS(originsOk, headersOk, methodsOk)(getRouter())))

	return srv.ListenAndServe()
	//return srv.ListenAndServeTLS("./certs/domain.crt", "./certs/domain.key")
}
