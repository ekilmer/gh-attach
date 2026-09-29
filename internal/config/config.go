package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Browsers         []BrowserEntry `yaml:"browsers"`
	SessionTokenFile string         `yaml:"session_token_file"`
}

type BrowserEntry struct {
	Browser         string `yaml:"browser"`
	Profile         string `yaml:"profile"`
	CookieStorePath string `yaml:"cookie_store_path"`
}

func defaultConfigDir() string {
	if xdg := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); xdg != "" {
		return filepath.Join(xdg, "gh")
	}

	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".config", "gh")
	}

	return "."
}

func DefaultConfigFile() string {
	return filepath.Join(defaultConfigDir(), "attach.yml")
}

func DefaultSessionTokenFile() string {
	dir := defaultConfigDir()
	if dir == "." {
		return ""
	}
	return filepath.Join(dir, "attach-session")
}

func LoadConfig(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, nil
		}
		return Config{}, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config file: %w", err)
	}

	for i := range cfg.Browsers {
		cfg.Browsers[i].Browser = strings.TrimSpace(cfg.Browsers[i].Browser)
		cfg.Browsers[i].Profile = strings.TrimSpace(cfg.Browsers[i].Profile)
		cfg.Browsers[i].CookieStorePath = strings.TrimSpace(cfg.Browsers[i].CookieStorePath)
	}
	cfg.SessionTokenFile = strings.TrimSpace(cfg.SessionTokenFile)

	return cfg, nil
}
