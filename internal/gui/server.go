package gui

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"api_hub/internal/config"
)

type Server struct {
	configPath string
	cfg        *config.Config
	manager    *RuntimeManager
	mux        *http.ServeMux
}

type ProviderState struct {
	Models []string `json:"models"`
}

type StateResponse struct {
	Defaults  config.DefaultsConfig    `json:"defaults"`
	Providers map[string]ProviderState `json:"providers"`
	Service   ServiceStatus            `json:"service"`
	Client    ClientInfo               `json:"client"`
}

type ClientInfo struct {
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
	Model   string `json:"model"`
}

func NewServer(configPath string, cfg *config.Config, manager *RuntimeManager) http.Handler {
	s := &Server{configPath: configPath, cfg: cfg, manager: manager, mux: http.NewServeMux()}
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/api/config", s.handleConfig)
	s.mux.HandleFunc("/api/defaults", s.handleDefaults)
	s.mux.HandleFunc("/api/service/start", s.handleStart)
	s.mux.HandleFunc("/api/service/stop", s.handleStop)
	s.mux.HandleFunc("/api/service/status", s.handleStatus)
	s.mux.HandleFunc("/api/test", s.handleTest)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(indexHTML))
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	s.writeState(w)
}

func (s *Server) handleDefaults(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if s.manager.Status().Running {
		writeJSONError(w, http.StatusBadRequest, "cannot change defaults while service is running")
		return
	}
	var defaults config.DefaultsConfig
	if err := json.NewDecoder(r.Body).Decode(&defaults); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	cfg, err := config.Load(s.configPath)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := cfg.ValidateDefaults(defaults); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := config.SaveDefaults(s.configPath, defaults); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeState(w)
}

func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err := s.manager.Start(); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeState(w)
}

func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err := s.manager.Stop(); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeState(w)
}

func (s *Server) handleTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var body struct {
		Provider string `json:"provider"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	cfg, err := config.Load(s.configPath)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	providerName := body.Provider
	if providerName == "" {
		providerName = cfg.Defaults.Provider
	}
	if providerName == "" {
		writeJSONError(w, http.StatusBadRequest, "no provider specified")
		return
	}
	provider, ok := cfg.Providers[providerName]
	if !ok {
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("provider %q not found", providerName))
		return
	}
	testURL := strings.TrimRight(provider.BaseURL, "/") + "/models"
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	if err != nil {
		writeJSON(w, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	req.Header.Set("Authorization", "Bearer "+provider.APIKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		writeJSON(w, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		writeJSON(w, map[string]any{"ok": true})
	} else {
		respBody, _ := io.ReadAll(resp.Body)
		writeJSON(w, map[string]any{"ok": false, "error": fmt.Sprintf("status %d: %s", resp.StatusCode, string(respBody))})
	}
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, s.manager.Status())
}

func (s *Server) writeState(w http.ResponseWriter) {
	cfg, err := config.Load(s.configPath)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	providers := make(map[string]ProviderState, len(cfg.Providers))
	for name, provider := range cfg.Providers {
		providers[name] = ProviderState{Models: append([]string(nil), provider.Models...)}
	}
	client := ClientInfo{
		BaseURL: "http://" + s.cfg.Server.Address + "/v1",
		APIKey:  s.cfg.Server.APIKey,
		Model:   cfg.Defaults.Model,
	}
	writeJSON(w, StateResponse{
		Defaults:  cfg.Defaults,
		Providers: providers,
		Service:   s.manager.Status(),
		Client:    client,
	})
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
