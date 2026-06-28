package config

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig              `yaml:"server"`
	Providers map[string]ProviderConfig `yaml:"providers"`
}

type ServerConfig struct {
	Address string `yaml:"address"`
	APIKey  string `yaml:"api_key"`
}

type ProviderConfig struct {
	BaseURL string   `yaml:"base_url"`
	APIKey  string   `yaml:"api_key"`
	Models  []string `yaml:"models"`
}

var providerNamePattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.Server.Address == "" {
		cfg.Server.Address = "127.0.0.1:8080"
	}
	cfg.Server.APIKey, err = expandEnv(cfg.Server.APIKey, "server.api_key")
	if err != nil {
		return nil, err
	}
	for name, provider := range cfg.Providers {
		provider.APIKey, err = expandEnv(provider.APIKey, fmt.Sprintf("providers.%s.api_key", name))
		if err != nil {
			return nil, err
		}
		cfg.Providers[name] = provider
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Validate() error {
	if strings.TrimSpace(c.Server.APIKey) == "" {
		return fmt.Errorf("server.api_key is required")
	}
	if len(c.Providers) == 0 {
		return fmt.Errorf("at least one provider is required")
	}
	for name, provider := range c.Providers {
		if !providerNamePattern.MatchString(name) {
			return fmt.Errorf("provider %q must contain only letters, numbers, underscore, or dash", name)
		}
		if strings.TrimSpace(provider.BaseURL) == "" {
			return fmt.Errorf("provider %q base_url is required", name)
		}
		parsed, err := url.Parse(provider.BaseURL)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return fmt.Errorf("provider %q base_url is invalid", name)
		}
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return fmt.Errorf("provider %q base_url must use http or https", name)
		}
		if strings.TrimSpace(provider.APIKey) == "" {
			return fmt.Errorf("provider %q api_key is required", name)
		}
	}
	return nil
}

func expandEnv(value, field string) (string, error) {
	if strings.HasPrefix(value, "${") {
		if !strings.HasSuffix(value, "}") {
			return "", fmt.Errorf("%s has malformed environment variable reference, expected ${NAME}", field)
		}
		name := strings.TrimSuffix(strings.TrimPrefix(value, "${"), "}")
		resolved := os.Getenv(name)
		if resolved == "" {
			return "", fmt.Errorf("%s references unset environment variable %s", field, name)
		}
		return resolved, nil
	}
	return value, nil
}
