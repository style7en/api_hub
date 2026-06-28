package models

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"api_hub/internal/config"
)

func TestHandlerReturnsPrefixedModels(t *testing.T) {
	cfg := &config.Config{Providers: map[string]config.ProviderConfig{
		"openai":   {Models: []string{"gpt-4o", "gpt-4o-mini"}},
		"deepseek": {Models: []string{"deepseek-chat"}},
	}}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)

	Handler(cfg).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	var body struct {
		Object string `json:"object"`
		Data   []struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			Created int64  `json:"created"`
			OwnedBy string `json:"owned_by"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Object != "list" {
		t.Fatalf("object = %q", body.Object)
	}
	wantModels := []struct {
		ID      string
		OwnedBy string
	}{
		{ID: "deepseek/deepseek-chat", OwnedBy: "deepseek"},
		{ID: "openai/gpt-4o", OwnedBy: "openai"},
		{ID: "openai/gpt-4o-mini", OwnedBy: "openai"},
	}
	if len(body.Data) != len(wantModels) {
		t.Fatalf("models length = %d, want %d", len(body.Data), len(wantModels))
	}
	for i, want := range wantModels {
		model := body.Data[i]
		if model.ID != want.ID {
			t.Fatalf("model[%d].id = %q, want %q", i, model.ID, want.ID)
		}
		if model.Object != "model" {
			t.Fatalf("model[%d].object = %q", i, model.Object)
		}
		if model.OwnedBy != want.OwnedBy {
			t.Fatalf("model[%d].owned_by = %q, want %q", i, model.OwnedBy, want.OwnedBy)
		}
		if model.Created != 0 {
			t.Fatalf("model[%d].created = %d", i, model.Created)
		}
	}
}

func TestHandlerReturnsEmptyDataArray(t *testing.T) {
	cfg := &config.Config{Providers: map[string]config.ProviderConfig{
		"openai": {Models: nil},
	}}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)

	Handler(cfg).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"data":[]`) {
		t.Fatalf("body = %s", rr.Body.String())
	}
	var body struct {
		Object string        `json:"object"`
		Data   []interface{} `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Object != "list" {
		t.Fatalf("object = %q", body.Object)
	}
	if len(body.Data) != 0 {
		t.Fatalf("data length = %d", len(body.Data))
	}
}
