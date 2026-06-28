package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadResolvesEnvironmentVariables(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-test")
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	err := os.WriteFile(path, []byte(`server:
  address: 127.0.0.1:8080
  api_key: local-key
providers:
  openai:
    base_url: https://api.openai.com/v1
    api_key: ${OPENAI_API_KEY}
    models:
      - gpt-4o
`), 0600)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.Server.Address != "127.0.0.1:8080" {
		t.Fatalf("address = %q", cfg.Server.Address)
	}
	if cfg.Server.APIKey != "local-key" {
		t.Fatalf("local api key = %q", cfg.Server.APIKey)
	}
	if cfg.Providers["openai"].APIKey != "sk-test" {
		t.Fatalf("provider api key = %q", cfg.Providers["openai"].APIKey)
	}
}

func TestLoadRejectsMissingProviderKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	err := os.WriteFile(path, []byte(`server:
  address: 127.0.0.1:8080
  api_key: local-key
providers:
  openai:
    base_url: https://api.openai.com/v1
    api_key: ${MISSING_API_KEY}
    models:
      - gpt-4o
`), 0600)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Load(path)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "MISSING_API_KEY") {
		t.Fatalf("error = %q, want MISSING_API_KEY", err)
	}
	if !strings.Contains(err.Error(), "providers.openai.api_key") {
		t.Fatalf("error = %q, want providers.openai.api_key", err)
	}
}

func TestLoadRejectsInvalidProviderName(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	err := os.WriteFile(path, []byte(`server:
  address: 127.0.0.1:8080
  api_key: local-key
providers:
  bad/name:
    base_url: https://api.example.com/v1
    api_key: sk-test
    models:
      - model
`), 0600)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Load(path)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "bad/name") {
		t.Fatalf("error = %q, want bad/name", err)
	}
}

func TestLoadRejectsUnsupportedBaseURLScheme(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	err := os.WriteFile(path, []byte(`server:
  address: 127.0.0.1:8080
  api_key: local-key
providers:
  openai:
    base_url: ftp://api.example.com/v1
    api_key: sk-test
    models:
      - model
`), 0600)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Load(path)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("error = %q, want base_url", err)
	}
	if !strings.Contains(err.Error(), "http or https") {
		t.Fatalf("error = %q, want http or https", err)
	}
}

func TestLoadRejectsMalformedBaseURL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	err := os.WriteFile(path, []byte(`server:
  address: 127.0.0.1:8080
  api_key: local-key
providers:
  openai:
    base_url: ://bad
    api_key: sk-test
    models:
      - model
`), 0600)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Load(path)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("error = %q, want base_url", err)
	}
	if !strings.Contains(err.Error(), "invalid") {
		t.Fatalf("error = %q, want invalid", err)
	}
}

func TestLoadRejectsMalformedProviderKeyEnvSyntax(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	err := os.WriteFile(path, []byte(`server:
  address: 127.0.0.1:8080
  api_key: local-key
providers:
  openai:
    base_url: https://api.openai.com/v1
    api_key: ${BROKEN
    models:
      - gpt-4o
`), 0600)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Load(path)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "providers.openai.api_key") {
		t.Fatalf("error = %q, want providers.openai.api_key", err)
	}
	if !strings.Contains(err.Error(), "malformed") {
		t.Fatalf("error = %q, want malformed", err)
	}
	if !strings.Contains(err.Error(), "${NAME}") {
		t.Fatalf("error = %q, want ${NAME}", err)
	}
}

func TestLoadReadsDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	err := os.WriteFile(path, []byte(`server:
  address: 127.0.0.1:8080
  api_key: local-key
defaults:
  provider: openrouter
  model: qwen/qwen3-coder:free
providers:
  openrouter:
    base_url: https://openrouter.ai/api/v1
    api_key: sk-test
    models:
      - qwen/qwen3-coder:free
`), 0600)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.Defaults.Provider != "openrouter" {
		t.Fatalf("default provider = %q", cfg.Defaults.Provider)
	}
	if cfg.Defaults.Model != "qwen/qwen3-coder:free" {
		t.Fatalf("default model = %q", cfg.Defaults.Model)
	}
}

func TestLoadRejectsInvalidDefaultProvider(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	err := os.WriteFile(path, []byte(`server:
  address: 127.0.0.1:8080
  api_key: local-key
defaults:
  provider: missing
  model: gpt-4o
providers:
  openai:
    base_url: https://api.openai.com/v1
    api_key: sk-test
    models:
      - gpt-4o
`), 0600)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Load(path)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "defaults.provider") {
		t.Fatalf("error = %v", err)
	}
}

func TestLoadRejectsInvalidDefaultModel(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	err := os.WriteFile(path, []byte(`server:
  address: 127.0.0.1:8080
  api_key: local-key
defaults:
  provider: openai
  model: missing-model
providers:
  openai:
    base_url: https://api.openai.com/v1
    api_key: sk-test
    models:
      - gpt-4o
`), 0600)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Load(path)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "defaults.model") {
		t.Fatalf("error = %v", err)
	}
}

func TestSaveDefaultsUpdatesYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	err := os.WriteFile(path, []byte(`server:
  address: 127.0.0.1:8080
  api_key: local-key
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

	if err := SaveDefaults(path, DefaultsConfig{Provider: "deepseek", Model: "deepseek-chat"}); err != nil {
		t.Fatalf("SaveDefaults returned error: %v", err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.Defaults.Provider != "deepseek" || cfg.Defaults.Model != "deepseek-chat" {
		t.Fatalf("defaults = %#v", cfg.Defaults)
	}
	if _, ok := cfg.Providers["openai"]; !ok {
		t.Fatal("openai provider was not preserved")
	}
}
