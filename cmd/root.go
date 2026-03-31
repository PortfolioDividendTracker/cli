package cmd

import (
	"fmt"
	"os"

	"github.com/PortfolioDividendTracker/cli/internal/config"
	"github.com/PortfolioDividendTracker/cli/internal/spec"
	"github.com/spf13/cobra"
)

var (
	flagToken string
	flagURL   string
)

var rootCmd = &cobra.Command{
	Use:           "pdt",
	Short:         "Portfolio Dividend Tracker CLI",
	Long:          "A CLI for AI agents to interact with the Portfolio Dividend Tracker API.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagToken, "token", "", "Personal Access Token (overrides PDT_TOKEN env and config)")
	rootCmd.PersistentFlags().StringVar(&flagURL, "url", "", "API base URL (overrides config)")
}

// NewRootCmd creates a fresh root command (used for testing).
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "pdt",
		Short:         "Portfolio Dividend Tracker CLI",
		Long:          "A CLI for AI agents to interact with the Portfolio Dividend Tracker API.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.PersistentFlags().StringVar(&flagToken, "token", "", "Personal Access Token (overrides PDT_TOKEN env and config)")
	cmd.PersistentFlags().StringVar(&flagURL, "url", "", "API base URL (overrides config)")
	return cmd
}

func Execute() error {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	baseURL := config.ResolveURL(flagURL, cfg)

	cachePath, err := spec.CachePath()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	if !spec.CacheExists(cachePath) {
		fmt.Fprintln(os.Stderr, "Fetching OpenAPI spec...")
		if err := spec.FetchAndCache(baseURL, cachePath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not fetch OpenAPI spec: %v\n", err)
			fmt.Fprintln(os.Stderr, "Run 'pdt update' after configuring the URL with 'pdt config set url <url>'")
		}
	}

	if spec.CacheExists(cachePath) {
		if err := RegisterDynamicCommands(rootCmd, cachePath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not load commands from spec: %v\n", err)
		}
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	return nil
}
