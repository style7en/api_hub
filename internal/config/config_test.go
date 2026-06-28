package config

import (
	"os"
	"path/filepath"
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
}
