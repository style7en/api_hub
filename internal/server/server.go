package server

import (
	"io"
	"net/http"

	"api-in-one/internal/auth"
	"api-in-one/internal/config"
	"api-in-one/internal/models"
	"api-in-one/internal/openaierror"
	"api-in-one/internal/proxy"
	"api-in-one/internal/router"
)

func New(cfg *config.Config) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/models", handleModels(cfg))
	mux.HandleFunc("/v1/", handleV1(cfg))
	return auth.Middleware(cfg.Server.APIKey, mux)
}

func handleModels(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			openaierror.Write(w, http.StatusMethodNotAllowed, "invalid_request_error", "method not allowed")
			return
		}
		models.Handler(cfg).ServeHTTP(w, r)
	}
}

func handleV1(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			openaierror.Write(w, http.StatusMethodNotAllowed, "invalid_request_error", "method not allowed")
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			openaierror.Write(w, http.StatusBadRequest, "invalid_request_error", "failed to read request body")
			return
		}
		providerName, rewrittenBody, err := router.RewriteModelBody(body)
		if err != nil {
			openaierror.Write(w, http.StatusBadRequest, "invalid_request_error", err.Error())
			return
		}
		provider, ok := cfg.Providers[providerName]
		if !ok {
			openaierror.Write(w, http.StatusBadRequest, "invalid_request_error", "unknown provider: "+providerName)
			return
		}
		if err := proxy.Forward(w, r, provider, rewrittenBody); err != nil {
			openaierror.Write(w, http.StatusBadGateway, "api_error", err.Error())
		}
	}
}
