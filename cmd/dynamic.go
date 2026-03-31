package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/PortfolioDividendTracker/cli/internal/client"
	"github.com/PortfolioDividendTracker/cli/internal/config"
	"github.com/PortfolioDividendTracker/cli/internal/spec"
	"github.com/spf13/cobra"
)

// RegisterDynamicCommands parses the cached OpenAPI spec and registers grouped Cobra commands.
// Operations are grouped by their OpenAPI tag into parent commands (e.g. "pdt bookings list").
func RegisterDynamicCommands(root *cobra.Command, specPath string) error {
	ops, err := spec.ParseOperations(specPath)
	if err != nil {
		return fmt.Errorf("failed to parse OpenAPI spec: %w", err)
	}

	groups := make(map[string]*cobra.Command)

	for _, op := range ops {
		op := op

		subCmd := &cobra.Command{
			Use:   op.SubCommand,
			Short: op.Summary,
			RunE:  makeRunFunc(op),
		}

		registered := make(map[string]bool)

		for _, p := range op.PathParams {
			subCmd.Flags().String(p.Name, "", p.Description)
			subCmd.MarkFlagRequired(p.Name)
			registered[p.Name] = true
		}

		for _, p := range op.QueryParams {
			if registered[p.Name] {
				continue
			}
			subCmd.Flags().String(p.Name, "", p.Description)
			if p.Required {
				subCmd.MarkFlagRequired(p.Name)
			}
			registered[p.Name] = true
		}

		if op.HasBody {
			subCmd.Flags().String("body", "", "Request body as JSON")
		}

		if op.Group == "" {
			root.AddCommand(subCmd)
			continue
		}

		groupCmd, exists := groups[op.Group]
		if !exists {
			groupCmd = &cobra.Command{
				Use:   op.Group,
				Short: fmt.Sprintf("Manage %s", op.Group),
			}
			groups[op.Group] = groupCmd
			root.AddCommand(groupCmd)
		}

		groupCmd.AddCommand(subCmd)
	}

	return nil
}

func makeRunFunc(op spec.Operation) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		token := config.ResolveToken(flagToken, cfg)
		baseURL := config.ResolveURL(flagURL, cfg)

		c := client.New(baseURL, token)

		pathParams := make(map[string]string)
		for _, p := range op.PathParams {
			val, _ := cmd.Flags().GetString(p.Name)
			pathParams[p.Name] = val
		}

		query := make(map[string]string)
		for _, p := range op.QueryParams {
			val, _ := cmd.Flags().GetString(p.Name)
			if val != "" {
				query[p.Name] = val
			}
		}

		var body []byte
		if op.HasBody {
			bodyStr, _ := cmd.Flags().GetString("body")
			if bodyStr != "" {
				body = []byte(bodyStr)
			}
		}

		result, statusCode, err := c.DoWithPathParams(op.Method, op.Path, pathParams, body, query)
		if err != nil {
			return err
		}

		if statusCode >= 400 {
			fmt.Fprintln(os.Stderr, string(result))
			return fmt.Errorf("HTTP %d", statusCode)
		}

		var prettyJSON json.RawMessage
		if err := json.Unmarshal(result, &prettyJSON); err != nil {
			fmt.Println(string(result))
			return nil
		}

		formatted, err := json.MarshalIndent(prettyJSON, "", "  ")
		if err != nil {
			fmt.Println(string(result))
			return nil
		}

		fmt.Println(string(formatted))
		return nil
	}
}
