package cmd

import (
	"fmt"
	"os"

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
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}
