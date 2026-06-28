package router

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"api_hub/internal/config"
)

func TestParseModelPrefix(t *testing.T) {
	provider, model, err := ParseModelPrefix("deepseek/deepseek-chat")
	if err != nil {
		t.Fatalf("ParseModelPrefix returned error: %v", err)
	}
	if provider != "deepseek" {
		t.Fatalf("provider = %q", provider)
	}
	if model != "deepseek-chat" {
		t.Fatalf("model = %q", model)
	}
}

func TestParseModelPrefixRejectsMissingPrefix(t *testing.T) {
	_, _, err := ParseModelPrefix("gpt-4o")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseModelPrefixRejectsWhitespaceAndEmptySegments(t *testing.T) {
	for _, model := range []string{" openai/gpt", "openai/ gpt", "/gpt", "openai/"} {
		_, _, err := ParseModelPrefix(model)
		if err == nil {
			t.Fatalf("ParseModelPrefix(%q) expected error", model)
		}
	}
}

func TestParseModelPrefixAllowsSlashInModel(t *testing.T) {
	provider, model, err := ParseModelPrefix("openai/folder/model")
	if err != nil {
		t.Fatalf("ParseModelPrefix returned error: %v", err)
	}
	if provider != "openai" {
		t.Fatalf("provider = %q", provider)
	}
	if model != "folder/model" {
		t.Fatalf("model = %q", model)
	}
}

func TestRewriteModelBody(t *testing.T) {
	body := []byte(`{"model":"openai/gpt-4o","messages":[{"role":"user","content":"hi"}],"stream":true}`)
	provider, rewritten, err := RewriteModelBody(body)
	if err != nil {
		t.Fatalf("RewriteModelBody returned error: %v", err)
	}
	if provider != "openai" {
		t.Fatalf("provider = %q", provider)
	}
	var decoded map[string]any
	if err := json.Unmarshal(rewritten, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["model"] != "gpt-4o" {
		t.Fatalf("model = %v", decoded["model"])
	}
	if decoded["stream"] != true {
		t.Fatalf("stream = %v", decoded["stream"])
	}
}

func TestRewriteModelBodyRejectsMissingModel(t *testing.T) {
	_, _, err := RewriteModelBody([]byte(`{"input":"hello"}`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRewriteModelBodyPreservesLargeInteger(t *testing.T) {
	body := []byte(`{"model":"openai/gpt-4o","seed":9007199254740993123}`)
	_, rewritten, err := RewriteModelBody(body)
	if err != nil {
		t.Fatalf("RewriteModelBody returned error: %v", err)
	}
	if !bytes.Contains(rewritten, []byte(`"seed":9007199254740993123`)) {
		t.Fatalf("rewritten body = %s", rewritten)
	}
}

func TestRewriteModelBodyRejectsInvalidJSON(t *testing.T) {
	_, _, err := RewriteModelBody([]byte(`{"model":"openai/gpt-4o"`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRewriteModelBodyRejectsNonStringModel(t *testing.T) {
	_, _, err := RewriteModelBody([]byte(`{"model":123}`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRewriteModelBodyRejectsEmptyModel(t *testing.T) {
	_, _, err := RewriteModelBody([]byte(`{"model":""}`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRewriteModelBodyWithDefaultsUsesUnprefixedDefault(t *testing.T) {
	body := []byte(`{"model":"anything","messages":[]}`)
	defaults := config.DefaultsConfig{Provider: "openrouter", Model: "qwen/qwen3-coder:free"}
	provider, rewritten, err := RewriteModelBodyWithDefaults(body, defaults)
	if err != nil {
		t.Fatalf("RewriteModelBodyWithDefaults returned error: %v", err)
	}
	if provider != "openrouter" {
		t.Fatalf("provider = %q", provider)
	}
	var decoded map[string]any
	if err := json.Unmarshal(rewritten, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["model"] != "qwen/qwen3-coder:free" {
		t.Fatalf("model = %v", decoded["model"])
	}
}

func TestRewriteModelBodyWithDefaultsKeepsPrefixedRouting(t *testing.T) {
	body := []byte(`{"model":"deepseek/deepseek-chat","messages":[]}`)
	defaults := config.DefaultsConfig{Provider: "openrouter", Model: "qwen/qwen3-coder:free"}
	provider, rewritten, err := RewriteModelBodyWithDefaults(body, defaults)
	if err != nil {
		t.Fatalf("RewriteModelBodyWithDefaults returned error: %v", err)
	}
	if provider != "deepseek" {
		t.Fatalf("provider = %q", provider)
	}
	var decoded map[string]any
	if err := json.Unmarshal(rewritten, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["model"] != "deepseek-chat" {
		t.Fatalf("model = %v", decoded["model"])
	}
}

func TestRewriteModelBodyWithDefaultsRejectsUnprefixedWithoutDefaults(t *testing.T) {
	_, _, err := RewriteModelBodyWithDefaults([]byte(`{"model":"gpt-4o"}`), config.DefaultsConfig{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "defaults") {
		t.Fatalf("error = %v", err)
	}
}

func TestForceModelReplacesClientModel(t *testing.T) {
	rewritten, err := ForceModel([]byte(`{"model":"whatever-client-sends","messages":[]}`), "gpt-4o")
	if err != nil {
		t.Fatalf("ForceModel returned error: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(rewritten, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["model"] != "gpt-4o" {
		t.Fatalf("model = %v", decoded["model"])
	}
}

func TestForceModelAddsModelFieldWhenMissing(t *testing.T) {
	rewritten, err := ForceModel([]byte(`{"input":"hello"}`), "text-embedding-3")
	if err != nil {
		t.Fatalf("ForceModel returned error: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(rewritten, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["model"] != "text-embedding-3" {
		t.Fatalf("model = %v", decoded["model"])
	}
}

func TestForceModelPreservesOtherFields(t *testing.T) {
	rewritten, err := ForceModel([]byte(`{"model":"old","seed":42}`), "new")
	if err != nil {
		t.Fatalf("ForceModel returned error: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(rewritten, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["model"] != "new" {
		t.Fatalf("model = %v", decoded["model"])
	}
	if decoded["seed"] != float64(42) {
		t.Fatalf("seed = %v", decoded["seed"])
	}
}
