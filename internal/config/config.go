package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	DefaultURL = "https://api.portfoliodividendtracker.com/v1"
	DirName    = ".pdt"
	FileName   = "config.json"
)

type Config struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

// Path returns the default config file path (~/.pdt/config.json).
func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, DirName, FileName), nil
}

// Load reads the config from the default path.
func Load() (*Config, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}
	return LoadFrom(path)
}

// LoadFrom reads the config from the given path. Returns defaults if the file does not exist.
func LoadFrom(path string) (*Config, error) {
	cfg := &Config{URL: DefaultURL}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	if cfg.URL == "" {
		cfg.URL = DefaultURL
	}

	return cfg, nil
}

// Save writes the config to the default path.
func Save(cfg *Config) error {
	path, err := Path()
	if err != nil {
		return err
	}
	return SaveTo(path, cfg)
}

// SaveTo writes the config to the given path, creating parent directories as needed.
func SaveTo(path string, cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// ResolveToken returns the token using priority: flag > env > config.
func ResolveToken(flag string, cfg *Config) string {
	if flag != "" {
		return flag
	}
	if env := os.Getenv("PDT_TOKEN"); env != "" {
		return env
	}
	return cfg.Token
}

// ResolveURL returns the URL using priority: flag > env > config.
func ResolveURL(flag string, cfg *Config) string {
	if flag != "" {
		return flag
	}
	if env := os.Getenv("PDT_URL"); env != "" {
		return env
	}
	return cfg.URL
}
