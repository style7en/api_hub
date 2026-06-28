package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"api-in-one/internal/config"
)

func TestForwardSendsProviderKeyAndBody(t *testing.T) {
	var gotPath string
	var gotAuth string
	var gotBody string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.String()
		gotAuth = r.Header.Get("Authorization")
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		gotBody = string(data)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer upstream.Close()

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions?x=1", strings.NewReader(`{"model":"gpt-4o"}`))
	rr := httptest.NewRecorder()

	err := Forward(rr, req, config.ProviderConfig{BaseURL: upstream.URL + "/v1", APIKey: "upstream-key"}, []byte(`{"model":"gpt-4o"}`))
	if err != nil {
		t.Fatalf("Forward returned error: %v", err)
	}
	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d", rr.Code)
	}
	if gotPath != "/v1/chat/completions?x=1" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotAuth != "Bearer upstream-key" {
		t.Fatalf("auth = %q", gotAuth)
	}
	if gotBody != `{"model":"gpt-4o"}` {
		t.Fatalf("body = %q", gotBody)
	}
	if rr.Body.String() != `{"ok":true}` {
		t.Fatalf("response body = %q", rr.Body.String())
	}
}

func TestForwardReturnsErrorOnNetworkFailure(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{}`))
	rr := httptest.NewRecorder()

	err := Forward(rr, req, config.ProviderConfig{BaseURL: "http://127.0.0.1:1/v1", APIKey: "upstream-key"}, []byte(`{}`))
	if err == nil {
		t.Fatal("expected error")
	}
}
