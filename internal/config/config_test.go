package config_test

import (
	"path/filepath"
	"testing"

	"github.com/PortfolioDividendTracker/cli/internal/config"
)

func TestLoadReturnsDefaultsWhenNoFile(t *testing.T) {
	dir := t.TempDir()
	cfg, err := config.LoadFrom(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.URL != "https://api.portfoliodividendtracker.com/v1" {
		t.Errorf("expected default URL, got %q", cfg.URL)
	}
	if cfg.Token != "" {
		t.Errorf("expected empty token, got %q", cfg.Token)
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := &config.Config{
		URL:   "https://custom.example.com/v1",
		Token: "pat_test123",
	}
	if err := config.SaveTo(path, cfg); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded, err := config.LoadFrom(path)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if loaded.URL != cfg.URL {
		t.Errorf("URL mismatch: got %q, want %q", loaded.URL, cfg.URL)
	}
	if loaded.Token != cfg.Token {
		t.Errorf("Token mismatch: got %q, want %q", loaded.Token, cfg.Token)
	}
}

func TestResolveTokenPriority(t *testing.T) {
	cfg := &config.Config{Token: "from_config"}

	if got := config.ResolveToken("", cfg); got != "from_config" {
		t.Errorf("expected config token, got %q", got)
	}

	t.Setenv("PDT_TOKEN", "from_env")
	if got := config.ResolveToken("", cfg); got != "from_env" {
		t.Errorf("expected env token, got %q", got)
	}

	if got := config.ResolveToken("from_flag", cfg); got != "from_flag" {
		t.Errorf("expected flag token, got %q", got)
	}
}

func TestResolveURLPriority(t *testing.T) {
	cfg := &config.Config{URL: "https://config.example.com/v1"}

	if got := config.ResolveURL("", cfg); got != "https://config.example.com/v1" {
		t.Errorf("expected config URL, got %q", got)
	}

	t.Setenv("PDT_URL", "https://env.example.com/v1")
	if got := config.ResolveURL("", cfg); got != "https://env.example.com/v1" {
		t.Errorf("expected env URL, got %q", got)
	}

	if got := config.ResolveURL("https://flag.example.com/v1", cfg); got != "https://flag.example.com/v1" {
		t.Errorf("expected flag URL, got %q", got)
	}
}
