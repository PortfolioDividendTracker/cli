package spec

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const CacheFileName = "openapi.json"

// CachePath returns the default spec cache path (~/.pdt/openapi.json).
func CachePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".pdt", CacheFileName), nil
}

// CacheExists returns true if the cached spec file exists.
func CacheExists(cachePath string) bool {
	_, err := os.Stat(cachePath)
	return err == nil
}

// FetchAndCache downloads the OpenAPI spec from baseURL/openapi.json and writes it to cachePath.
// The baseURL should be the API base URL (e.g. https://api.example.com/v1).
// The spec is fetched from the root domain, not the versioned path.
func FetchAndCache(baseURL string, cachePath string) error {
	specURL := strings.TrimSuffix(baseURL, "/v1") + "/openapi.json"

	resp, err := http.Get(specURL)
	if err != nil {
		return fmt.Errorf("failed to fetch OpenAPI spec: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch OpenAPI spec: HTTP %d", resp.StatusCode)
	}

	if err := os.MkdirAll(filepath.Dir(cachePath), 0700); err != nil {
		return err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read OpenAPI spec: %w", err)
	}

	return os.WriteFile(cachePath, data, 0600)
}
