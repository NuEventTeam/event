package chat

import (
	"context"
	"github.com/NuEventTeam/events/pkg"
	"log"
	"net/http"
)

func Authorize(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		log.Println("TRYING TO CHATT")
		userId, userAgent, err := pkg.ParseJWT(token, "my-32-character-ultra-secure-and-ultra-long-secret")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Malformed Token"))
			log.Println(err)
			return
		}

		ctx := context.WithValue(r.Context(), "userId", userId)
		ctx = context.WithValue(ctx, "userAgent", userAgent)
		next.ServeHTTP(w, r.WithContext(ctx))

	})

}
