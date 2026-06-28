package gui

import (
	"context"
	"net/http"
	"sync"
	"time"

	"api_hub/internal/config"
	"api_hub/internal/server"
)

type ServiceStatus struct {
	Running bool     `json:"running"`
	Logs    []string `json:"logs"`
}

type RuntimeManager struct {
	mu         sync.Mutex
	configPath string
	cfg        *config.Config
	srv        *http.Server
	logs       []string
}

func NewRuntimeManager(configPath string, cfg *config.Config) *RuntimeManager {
	return &RuntimeManager{configPath: configPath, cfg: cfg, logs: make([]string, 0)}
}

func (m *RuntimeManager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.srv != nil {
		return nil
	}
	cfg, err := config.Load(m.configPath)
	if err != nil {
		return err
	}
	m.cfg = cfg
	handler := server.New(cfg)
	srv := &http.Server{Addr: m.cfg.Server.Address, Handler: handler}
	m.srv = srv
	m.addLogLocked("API server starting on " + m.cfg.Server.Address)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			m.addLog("API server error: " + err.Error())
		}
	}()
	return nil
}

func (m *RuntimeManager) Stop() error {
	m.mu.Lock()
	srv := m.srv
	m.srv = nil
	m.mu.Unlock()
	if srv == nil {
		return nil
	}
	m.addLog("API server stopping")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(ctx)
}

func (m *RuntimeManager) Status() ServiceStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	logs := append([]string(nil), m.logs...)
	return ServiceStatus{Running: m.srv != nil, Logs: logs}
}

func (m *RuntimeManager) Config() *config.Config {
	return m.cfg
}

func (m *RuntimeManager) addLog(line string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.addLogLocked(line)
}

func (m *RuntimeManager) addLogLocked(line string) {
	m.logs = append(m.logs, line)
	if len(m.logs) > 200 {
		m.logs = m.logs[len(m.logs)-200:]
	}
}
