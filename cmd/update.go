package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/PortfolioDividendTracker/cli/internal/config"
	"github.com/PortfolioDividendTracker/cli/internal/spec"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Fetch the latest OpenAPI spec from the API",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		baseURL := config.ResolveURL(flagURL, cfg)

		cachePath, err := spec.CachePath()
		if err != nil {
			return err
		}

		fmt.Fprintln(os.Stderr, "Fetching OpenAPI spec...")
		if err := spec.FetchAndCache(baseURL, cachePath); err != nil {
			return fmt.Errorf("failed to update: %w", err)
		}

		return json.NewEncoder(os.Stdout).Encode(map[string]string{
			"status": "ok",
			"cached": cachePath,
		})
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
