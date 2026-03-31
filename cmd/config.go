package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/PortfolioDividendTracker/cli/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a config value (url, token)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		switch args[0] {
		case "url":
			cfg.URL = args[1]
		case "token":
			cfg.Token = args[1]
		default:
			return fmt.Errorf("unknown config key: %s (valid keys: url, token)", args[0])
		}

		if err := config.Save(cfg); err != nil {
			return err
		}

		return json.NewEncoder(os.Stdout).Encode(map[string]string{"status": "ok", "key": args[0]})
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a config value (url, token)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		var value string
		switch args[0] {
		case "url":
			value = cfg.URL
		case "token":
			value = cfg.Token
		default:
			return fmt.Errorf("unknown config key: %s (valid keys: url, token)", args[0])
		}

		return json.NewEncoder(os.Stdout).Encode(map[string]string{"key": args[0], "value": value})
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	rootCmd.AddCommand(configCmd)
}
