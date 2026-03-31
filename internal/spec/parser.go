package spec

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/getkin/kin-openapi/openapi3"
)

type Param struct {
	Name        string
	Description string
	Required    bool
	SchemaType  string
}

type Operation struct {
	OperationID string
	CommandName string
	Group       string
	Tag         string
	SubCommand  string
	Method      string
	Path        string
	Summary     string
	PathParams  []Param
	QueryParams []Param
	HasBody     bool
}

type ParseResult struct {
	Operations      []Operation
	TagDescriptions map[string]string // tag name → description
}

// ParseOperations reads an OpenAPI spec file and returns all operations and tag descriptions.
func ParseOperations(path string) (*ParseResult, error) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(path)
	if err != nil {
		return nil, err
	}

	// Skip validation — real-world specs may have quirks and we only need paths/operations.

	// Build tag description map
	tagDescs := make(map[string]string)
	for _, tag := range doc.Tags {
		if tag.Description != "" {
			tagDescs[tag.Name] = tag.Description
		}
	}

	var ops []Operation

	for path, pathItem := range doc.Paths.Map() {
		for method, op := range pathItem.Operations() {
			if op.OperationID == "" {
				continue
			}

			cmdName := OperationIDToCommandName(op.OperationID)
			tag := ""
			if len(op.Tags) > 0 {
				tag = op.Tags[0]
			}
			group, subCmd := TagToGroupAndSubCommand(op.Tags, cmdName)

			operation := Operation{
				OperationID: op.OperationID,
				CommandName: cmdName,
				Group:       group,
				Tag:         tag,
				SubCommand:  subCmd,
				Method:      strings.ToUpper(method),
				Path:        path,
				Summary:     op.Summary,
				HasBody:     op.RequestBody != nil,
			}

			for _, paramRef := range op.Parameters {
				param := paramRef.Value
				if param == nil {
					continue
				}

				p := Param{
					Name:        param.Name,
					Description: param.Description,
					Required:    param.Required,
				}

				if param.Schema != nil && param.Schema.Value != nil {
					types := param.Schema.Value.Type.Slice()
					if len(types) > 0 {
						p.SchemaType = types[0]
					}
				}

				switch param.In {
				case "path":
					operation.PathParams = append(operation.PathParams, p)
				case "query":
					operation.QueryParams = append(operation.QueryParams, p)
				}
			}

			ops = append(ops, operation)
		}
	}

	return &ParseResult{
		Operations:      ops,
		TagDescriptions: tagDescs,
	}, nil
}

var camelSplitter = regexp.MustCompile(`([a-z])([A-Z])`)

// OperationIDToCommandName converts an operationId like "booking.listBookingsEndpoint"
// to a CLI command name like "list-bookings".
func OperationIDToCommandName(operationID string) string {
	// Remove domain prefix (e.g. "booking." from "booking.listBookingsEndpoint")
	if idx := strings.LastIndex(operationID, "."); idx != -1 {
		operationID = operationID[idx+1:]
	}

	// Remove "Endpoint" suffix
	operationID = strings.TrimSuffix(operationID, "Endpoint")

	// Split camelCase into kebab-case
	result := camelSplitter.ReplaceAllString(operationID, "${1}-${2}")

	// Lowercase everything
	result = strings.Map(func(r rune) rune {
		if unicode.IsUpper(r) {
			return unicode.ToLower(r)
		}
		return r
	}, result)

	return result
}

// TagToGroupAndSubCommand extracts a CLI group and subcommand from an OpenAPI tag and command name.
// For example: tag "User → Bookings" with command "list-bookings" → group "bookings", subcommand "list".
// Tag "Portfolio → Gains" with command "get-portfolio-gains" → group "portfolio", subcommand "gains".
// Tag "Portfolio" with command "get-portfolio" → group "portfolio", subcommand "get".
func TagToGroupAndSubCommand(tags []string, commandName string) (string, string) {
	if len(tags) == 0 {
		return "", commandName
	}

	tag := tags[0]

	// Split tag into segments: "User → Bookings" → ["User", "Bookings"]
	segments := strings.Split(tag, "→")
	for i := range segments {
		segments[i] = strings.TrimSpace(segments[i])
	}

	// The last segment becomes the group name
	lastSegment := segments[len(segments)-1]
	group := segmentToKebab(lastSegment)

	// Collect all tag segments as words to strip from the command name
	var stripWords []string
	for _, seg := range segments {
		kebab := segmentToKebab(seg)
		stripWords = append(stripWords, strings.Split(kebab, "-")...)
	}

	subCmd := stripWordsFromCommand(commandName, stripWords)

	return group, subCmd
}

func segmentToKebab(segment string) string {
	result := camelSplitter.ReplaceAllString(segment, "${1}-${2}")
	result = strings.Map(func(r rune) rune {
		if r == ' ' {
			return '-'
		}
		if unicode.IsUpper(r) {
			return unicode.ToLower(r)
		}
		return r
	}, result)
	return result
}

// stripWordsFromCommand removes tag-related words from a command name to derive the subcommand.
// Handles singular/plural matching (e.g. "booking" matches "bookings" and vice versa).
func stripWordsFromCommand(commandName string, words []string) string {
	cmdParts := strings.Split(commandName, "-")

	wordSet := make(map[string]bool)
	for _, w := range words {
		wordSet[w] = true
		wordSet[strings.TrimSuffix(w, "s")] = true
		wordSet[w+"s"] = true
	}

	var result []string
	for _, cp := range cmdParts {
		if wordSet[cp] || wordSet[strings.TrimSuffix(cp, "s")] {
			continue
		}
		result = append(result, cp)
	}

	if len(result) == 0 {
		return cmdParts[0]
	}

	return strings.Join(result, "-")
}
