package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/PortfolioDividendTracker/cli/internal/client"
	"github.com/PortfolioDividendTracker/cli/internal/config"
	"github.com/PortfolioDividendTracker/cli/internal/spec"
	"github.com/spf13/cobra"
)

func firstSentence(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, ". "); i != -1 {
		return s[:i+1]
	}
	if i := strings.Index(s, ".\n"); i != -1 {
		return s[:i+1]
	}
	// Truncate at first newline if no period
	if i := strings.IndexByte(s, '\n'); i != -1 {
		return s[:i]
	}
	return s
}

// RegisterDynamicCommands parses the cached OpenAPI spec and registers grouped Cobra commands.
// Operations are grouped by their OpenAPI tag into parent commands (e.g. "pdt bookings list").
func RegisterDynamicCommands(root *cobra.Command, specPath string) error {
	result, err := spec.ParseOperations(specPath)
	if err != nil {
		return fmt.Errorf("failed to parse OpenAPI spec: %w", err)
	}

	groups := make(map[string]*cobra.Command)

	for _, op := range result.Operations {
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
			groupDesc := firstSentence(result.TagDescriptions[op.Tag])
			if groupDesc == "" || strings.HasPrefix(groupDesc, "#") {
				groupDesc = fmt.Sprintf("Manage %s", op.Group)
			}
			groupCmd = &cobra.Command{
				Use:   op.Group,
				Short: groupDesc,
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
			return formatAPIError(result, statusCode)
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

func formatAPIError(body []byte, statusCode int) error {
	var apiErr struct {
		Message string            `json:"message"`
		Errors  map[string][]string `json:"errors"`
	}

	if err := json.Unmarshal(body, &apiErr); err != nil || apiErr.Message == "" {
		// Can't parse — show raw body
		fmt.Fprintln(os.Stderr, string(body))
		return fmt.Errorf("HTTP %d", statusCode)
	}

	switch statusCode {
	case 401:
		return fmt.Errorf("authentication failed: %s\nCheck your token with: pdt config get token", apiErr.Message)
	case 403:
		return fmt.Errorf("access denied: %s", apiErr.Message)
	case 404:
		return fmt.Errorf("not found: %s", apiErr.Message)
	case 422:
		msg := fmt.Sprintf("validation failed: %s", apiErr.Message)
		for field, errs := range apiErr.Errors {
			for _, e := range errs {
				msg += fmt.Sprintf("\n  %s: %s", field, e)
			}
		}
		return fmt.Errorf("%s", msg)
	case 429:
		return fmt.Errorf("rate limited: %s", apiErr.Message)
	default:
		return fmt.Errorf("error %d: %s", statusCode, apiErr.Message)
	}
}
