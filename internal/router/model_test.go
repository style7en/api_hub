package router

import (
	"bytes"
	"encoding/json"
	"testing"
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
