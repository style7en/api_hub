package gui

import (
	"testing"

	"api-in-one/internal/config"
)

func TestRuntimeManagerStartStop(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{Address: "127.0.0.1:0", APIKey: "test-key"},
		Providers: map[string]config.ProviderConfig{
			"openai": {BaseURL: "https://api.openai.com/v1", APIKey: "sk-test", Models: []string{"gpt-4o"}},
		},
	}
	manager := NewRuntimeManager(cfg)
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

func TestRuntimeManagerAddsLifecycleLogs(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{Address: "127.0.0.1:0", APIKey: "test-key"},
		Providers: map[string]config.ProviderConfig{
			"openai": {BaseURL: "https://api.openai.com/v1", APIKey: "sk-test", Models: []string{"gpt-4o"}},
		},
	}
	manager := NewRuntimeManager(cfg)
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
