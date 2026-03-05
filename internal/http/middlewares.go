package http

import (
	"net/http"
	"os"
	"strings"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			errDTO := NewErrorDTO(ErrTokenEmpty)
			http.Error(w, errDTO.String(), http.StatusUnauthorized)
			return
		}

		token := strings.Split(authHeader, " ")[1]

		if token != os.Getenv("AUTH_TOKEN") {
			errDTO := NewErrorDTO(ErrInvalidToken)
			http.Error(w, errDTO.String(), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
