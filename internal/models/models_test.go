package models

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"api-in-one/internal/config"
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
			ID     string `json:"id"`
			Object string `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Object != "list" {
		t.Fatalf("object = %q", body.Object)
	}
	ids := map[string]bool{}
	for _, model := range body.Data {
		ids[model.ID] = true
		if model.Object != "model" {
			t.Fatalf("model object = %q", model.Object)
		}
	}
	for _, want := range []string{"openai/gpt-4o", "openai/gpt-4o-mini", "deepseek/deepseek-chat"} {
		if !ids[want] {
			t.Fatalf("missing model %q in %#v", want, ids)
		}
	}
}
