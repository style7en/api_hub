package gui

import (
	"os"
	"path/filepath"
	"testing"

	"api_hub/internal/config"
)

func TestRuntimeManagerStartStop(t *testing.T) {
	path := writeRuntimeConfig(t)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	manager := NewRuntimeManager(path, cfg)
	if manager.Status().Running {
		t.Fatal("expected stopped initially")
	}
	if err := manager.Start(); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	if !manager.Status().Running {
		t.Fatal("expected running after start")
	}
	if err := manager.Start(); err != nil {
		t.Fatalf("second Start returned error: %v", err)
	}
	if err := manager.Stop(); err != nil {
		t.Fatalf("Stop returned error: %v", err)
	}
	if manager.Status().Running {
		t.Fatal("expected stopped after stop")
	}
}

func TestRuntimeManagerReloadsConfigOnStart(t *testing.T) {
	path := writeRuntimeConfig(t)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	manager := NewRuntimeManager(path, cfg)

	config.SaveDefaults(path, config.DefaultsConfig{Provider: "openai", Model: "gpt-4o"})
	if err := manager.Start(); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	defer manager.Stop()

	if manager.Config().Defaults.Model != "gpt-4o" {
		t.Fatalf("defaults.model = %q, expected gpt-4o", manager.Config().Defaults.Model)
	}
}

func TestRuntimeManagerAddsLifecycleLogs(t *testing.T) {
	path := writeRuntimeConfig(t)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	manager := NewRuntimeManager(path, cfg)
	if err := manager.Start(); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	if err := manager.Stop(); err != nil {
		t.Fatalf("Stop returned error: %v", err)
	}
	status := manager.Status()
	if len(status.Logs) == 0 {
		t.Fatal("expected lifecycle logs")
	}
}

func writeRuntimeConfig(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	err := os.WriteFile(path, []byte(`server:
  address: 127.0.0.1:0
  api_key: local-key
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
	return path
}
