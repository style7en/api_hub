package gui

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"api_hub/internal/config"
)

func TestServerReturnsConfigState(t *testing.T) {
	path := writeGUIConfig(t)
	cfg, _ := config.Load(path)
	server := NewServer(path, cfg, NewRuntimeManager(path, cfg))
	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	rr := httptest.NewRecorder()

	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", rr.Code, rr.Body.String())
	}
	var state StateResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	if state.Defaults.Provider != "openai" {
		t.Fatalf("defaults = %#v", state.Defaults)
	}
	if len(state.Providers["openai"].Models) != 1 {
		t.Fatalf("providers = %#v", state.Providers)
	}
	if state.Client.BaseURL == "" {
		t.Fatal("expected client base_url")
	}
	if state.Client.APIKey == "" {
		t.Fatal("expected client api_key")
	}
}

func TestServerSavesDefaults(t *testing.T) {
	path := writeGUIConfig(t)
	cfg, _ := config.Load(path)
	server := NewServer(path, cfg, NewRuntimeManager(path, cfg))
	body := bytes.NewBufferString(`{"provider":"deepseek","model":"deepseek-chat"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/defaults", body)
	rr := httptest.NewRecorder()

	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", rr.Code, rr.Body.String())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "provider: deepseek") || !strings.Contains(string(data), "model: deepseek-chat") {
		t.Fatalf("config = %s", string(data))
	}
}

func TestServerRejectsDefaultsWhileRunning(t *testing.T) {
	path := writeGUIConfig(t)
	cfg, _ := config.Load(path)
	manager := NewRuntimeManager(path, cfg)
	manager.Start()
	defer manager.Stop()
	server := NewServer(path, cfg, manager)
	body := bytes.NewBufferString(`{"provider":"deepseek","model":"deepseek-chat"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/defaults", body)
	rr := httptest.NewRecorder()

	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "cannot change defaults while service is running") {
		t.Fatalf("body = %s", rr.Body.String())
	}
}

func TestServerRejectsInvalidDefaults(t *testing.T) {
	path := writeGUIConfig(t)
	cfg, _ := config.Load(path)
	server := NewServer(path, cfg, NewRuntimeManager(path, cfg))
	body := bytes.NewBufferString(`{"provider":"missing","model":"x"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/defaults", body)
	rr := httptest.NewRecorder()

	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rr.Code)
	}
}

func TestServerStartStopStatus(t *testing.T) {
	path := writeGUIConfig(t)
	cfg, _ := config.Load(path)
	manager := NewRuntimeManager(path, cfg)
	server := NewServer(path, cfg, manager)

	startReq := httptest.NewRequest(http.MethodPost, "/api/service/start", nil)
	startRR := httptest.NewRecorder()
	server.ServeHTTP(startRR, startReq)
	if startRR.Code != http.StatusOK {
		t.Fatalf("start status = %d body = %s", startRR.Code, startRR.Body.String())
	}
	if !manager.Status().Running {
		t.Fatal("expected process running")
	}

	stopReq := httptest.NewRequest(http.MethodPost, "/api/service/stop", nil)
	stopRR := httptest.NewRecorder()
	server.ServeHTTP(stopRR, stopReq)
	if stopRR.Code != http.StatusOK {
		t.Fatalf("stop status = %d body = %s", stopRR.Code, stopRR.Body.String())
	}
	if manager.Status().Running {
		t.Fatal("expected process stopped")
	}
}

func TestServerReturnsHTMLPage(t *testing.T) {
	path := writeGUIConfig(t)
	cfg, _ := config.Load(path)
	server := NewServer(path, cfg, NewRuntimeManager(path, cfg))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "API Hub") {
		t.Fatalf("body = %s", rr.Body.String())
	}
}

func writeGUIConfig(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	err := os.WriteFile(path, []byte(`server:
  address: 127.0.0.1:0
  api_key: local-key
defaults:
  provider: openai
  model: gpt-4o
providers:
  openai:
    base_url: https://api.openai.com/v1
    api_key: sk-test
    models:
      - gpt-4o
  deepseek:
    base_url: https://api.deepseek.com/v1
    api_key: sk-deepseek
    models:
      - deepseek-chat
`), 0600)
	if err != nil {
		t.Fatal(err)
	}
	return path
}
