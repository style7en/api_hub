package auth

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"api-in-one/internal/openaierror"
)

func Middleware(apiKey string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if subtle.ConstantTimeCompare([]byte(got), []byte(apiKey)) != 1 {
			openaierror.Write(w, http.StatusUnauthorized, "authentication_error", "invalid or missing bearer token")
			return
		}
		next.ServeHTTP(w, r)
	})
}
