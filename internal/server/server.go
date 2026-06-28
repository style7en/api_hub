package server

import (
	"io"
	"net/http"

	"api_hub/internal/auth"
	"api_hub/internal/config"
	"api_hub/internal/models"
	"api_hub/internal/openaierror"
	"api_hub/internal/proxy"
	"api_hub/internal/router"
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
		if cfg.Defaults.Provider == "" || cfg.Defaults.Model == "" {
			openaierror.Write(w, http.StatusBadRequest, "invalid_request_error", "no provider selected, use GUI to select one")
			return
		}
		provider, ok := cfg.Providers[cfg.Defaults.Provider]
		if !ok {
			openaierror.Write(w, http.StatusBadRequest, "invalid_request_error", "unknown provider: "+cfg.Defaults.Provider)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			openaierror.Write(w, http.StatusBadRequest, "invalid_request_error", "failed to read request body")
			return
		}
		rewrittenBody, err := router.ForceModel(body, cfg.Defaults.Model)
		if err != nil {
			openaierror.Write(w, http.StatusBadRequest, "invalid_request_error", err.Error())
			return
		}
		if err := proxy.Forward(w, r, provider, rewrittenBody); err != nil {
			openaierror.Write(w, http.StatusBadGateway, "api_error", err.Error())
		}
	}
}
