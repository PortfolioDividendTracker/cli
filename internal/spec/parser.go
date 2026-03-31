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
	Method      string
	Path        string
	Summary     string
	PathParams  []Param
	QueryParams []Param
	HasBody     bool
}

// ParseOperations reads an OpenAPI spec file and returns all operations as a slice.
func ParseOperations(path string) ([]Operation, error) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(path)
	if err != nil {
		return nil, err
	}

	// Skip validation — real-world specs may have quirks and we only need paths/operations.

	var ops []Operation

	for path, pathItem := range doc.Paths.Map() {
		for method, op := range pathItem.Operations() {
			if op.OperationID == "" {
				continue
			}

			operation := Operation{
				OperationID: op.OperationID,
				CommandName: OperationIDToCommandName(op.OperationID),
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

	return ops, nil
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
