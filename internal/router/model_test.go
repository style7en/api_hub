package router

import (
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
