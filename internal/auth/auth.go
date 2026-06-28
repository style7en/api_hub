package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"strings"

	"api-in-one/internal/openaierror"
)

func Middleware(apiKey string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")
		if !strings.HasPrefix(authorization, "Bearer ") {
			openaierror.Write(w, http.StatusUnauthorized, "authentication_error", "invalid or missing bearer token")
			return
		}
		got := strings.TrimPrefix(authorization, "Bearer ")
		gotDigest := sha256.Sum256([]byte(got))
		apiKeyDigest := sha256.Sum256([]byte(apiKey))
		if subtle.ConstantTimeCompare(gotDigest[:], apiKeyDigest[:]) != 1 {
			openaierror.Write(w, http.StatusUnauthorized, "authentication_error", "invalid or missing bearer token")
			return
		}
		next.ServeHTTP(w, r)
	})
}
