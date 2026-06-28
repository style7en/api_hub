package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"api-in-one/internal/config"
)

func TestServerRoutesChatCompletionByModelPrefix(t *testing.T) {
	var gotBody string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		gotBody = string(data)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"chatcmpl-test"}`))
	}))
	defer upstream.Close()

	cfg := testConfig(upstream.URL)
	handler := New(cfg)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"openai/gpt-4o","messages":[]}`))
	req.Header.Set("Authorization", "Bearer local-key")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", rr.Code, rr.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal([]byte(gotBody), &body); err != nil {
		t.Fatal(err)
	}
	if body["model"] != "gpt-4o" {
		t.Fatalf("upstream model = %v", body["model"])
	}
}

func TestServerRejectsUnknownProvider(t *testing.T) {
	cfg := testConfig("http://127.0.0.1:1")
	handler := New(cfg)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"missing/gpt-4o"}`))
	req.Header.Set("Authorization", "Bearer local-key")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "unknown provider") {
		t.Fatalf("body = %s", rr.Body.String())
	}
}

func TestServerRejectsModelLessTransparentRequest(t *testing.T) {
	cfg := testConfig("http://127.0.0.1:1")
	handler := New(cfg)
	req := httptest.NewRequest(http.MethodPost, "/v1/audio/transcriptions", strings.NewReader(`{"file":"x"}`))
	req.Header.Set("Authorization", "Bearer local-key")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "request body must include model") {
		t.Fatalf("body = %s", rr.Body.String())
	}
}

func TestServerModelsRejectsPost(t *testing.T) {
	cfg := testConfig("http://127.0.0.1:1")
	handler := New(cfg)
	req := httptest.NewRequest(http.MethodPost, "/v1/models", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer local-key")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "method not allowed") {
		t.Fatalf("body = %s", rr.Body.String())
	}
}

func TestServerChatRejectsGet(t *testing.T) {
	cfg := testConfig("http://127.0.0.1:1")
	handler := New(cfg)
	req := httptest.NewRequest(http.MethodGet, "/v1/chat/completions", nil)
	req.Header.Set("Authorization", "Bearer local-key")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "method not allowed") {
		t.Fatalf("body = %s", rr.Body.String())
	}
}

func TestServerDoesNotAppendErrorAfterPartialUpstreamResponse(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", "100")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"partial":`))
	}))
	defer upstream.Close()

	cfg := testConfig(upstream.URL)
	handler := New(cfg)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"openai/gpt-4o","messages":[]}`))
	req.Header.Set("Authorization", "Bearer local-key")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if strings.Contains(rr.Body.String(), "api_error") {
		t.Fatalf("body contains appended api_error: %s", rr.Body.String())
	}
}

func TestServerRequiresAuth(t *testing.T) {
	cfg := testConfig("http://127.0.0.1:1")
	handler := New(cfg)
	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rr.Code)
	}
}

func TestServerPassesThroughStreamingResponse(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.(http.Flusher).Flush()
		_, _ = w.Write([]byte("data: {\"delta\":\"hi\"}\n\n"))
		w.(http.Flusher).Flush()
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer upstream.Close()

	cfg := testConfig(upstream.URL)
	handler := New(cfg)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"openai/gpt-4o","stream":true}`))
	req.Header.Set("Authorization", "Bearer local-key")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	if rr.Header().Get("Content-Type") != "text/event-stream" {
		t.Fatalf("content-type = %q", rr.Header().Get("Content-Type"))
	}
	if !strings.Contains(rr.Body.String(), "data: [DONE]") {
		t.Fatalf("body = %s", rr.Body.String())
	}
}

func testConfig(baseURL string) *config.Config {
	return &config.Config{
		Server: config.ServerConfig{Address: "127.0.0.1:0", APIKey: "local-key"},
		Providers: map[string]config.ProviderConfig{
			"openai": {BaseURL: baseURL + "/v1", APIKey: "upstream-key", Models: []string{"gpt-4o"}},
		},
	}
}
