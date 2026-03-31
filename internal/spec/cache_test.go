package spec_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/PortfolioDividendTracker/cli/internal/spec"
)

func TestFetchAndCache(t *testing.T) {
	specJSON := `{"openapi":"3.1.0","info":{"title":"Test","version":"1.0"},"paths":{}}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/openapi.json" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(specJSON))
	}))
	defer server.Close()

	cacheDir := t.TempDir()
	cachePath := filepath.Join(cacheDir, "openapi.json")

	err := spec.FetchAndCache(server.URL, cachePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("cache file not created: %v", err)
	}
	if string(data) != specJSON {
		t.Errorf("cached content mismatch: got %q", string(data))
	}
}

func TestFetchAndCacheHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cachePath := filepath.Join(t.TempDir(), "openapi.json")

	err := spec.FetchAndCache(server.URL, cachePath)
	if err == nil {
		t.Fatal("expected error for HTTP 500")
	}
}

func TestCachePathDefault(t *testing.T) {
	path, err := spec.CachePath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filepath.Base(path) != "openapi.json" {
		t.Errorf("expected openapi.json, got %s", filepath.Base(path))
	}
}

func TestCacheExists(t *testing.T) {
	dir := t.TempDir()

	if spec.CacheExists(filepath.Join(dir, "nonexistent.json")) {
		t.Error("expected false for nonexistent file")
	}

	path := filepath.Join(dir, "exists.json")
	os.WriteFile(path, []byte("{}"), 0600)
	if !spec.CacheExists(path) {
		t.Error("expected true for existing file")
	}
}
