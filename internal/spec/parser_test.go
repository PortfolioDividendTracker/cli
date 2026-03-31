package spec_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/PortfolioDividendTracker/cli/internal/spec"
)

func TestParseOperations(t *testing.T) {
	specJSON := `{
		"openapi": "3.1.0",
		"info": {"title": "Test", "version": "1.0"},
		"paths": {
			"/bookings": {
				"get": {
					"operationId": "booking.listBookingsEndpoint",
					"summary": "List bookings",
					"parameters": [
						{"name": "page", "in": "query", "schema": {"type": "integer"}, "description": "Page number"},
						{"name": "perPage", "in": "query", "schema": {"type": "integer"}, "description": "Items per page"}
					]
				},
				"post": {
					"operationId": "booking.createBookingEndpoint",
					"summary": "Create a booking",
					"parameters": [
						{"name": "brokerId", "in": "query", "required": true, "schema": {"type": "integer"}, "description": "Broker ID"},
						{"name": "date", "in": "query", "required": true, "schema": {"type": "string"}, "description": "Booking date"}
					]
				}
			},
			"/bookings/{bookingId}": {
				"get": {
					"operationId": "booking.getBookingEndpoint",
					"summary": "Get a booking",
					"parameters": [
						{"name": "bookingId", "in": "path", "required": true, "schema": {"type": "integer"}, "description": "Booking ID"}
					]
				},
				"delete": {
					"operationId": "booking.deleteBookingEndpoint",
					"summary": "Delete a booking",
					"parameters": [
						{"name": "bookingId", "in": "path", "required": true, "schema": {"type": "integer"}, "description": "Booking ID"}
					]
				}
			}
		}
	}`

	path := filepath.Join(t.TempDir(), "openapi.json")
	os.WriteFile(path, []byte(specJSON), 0600)

	result, err := spec.ParseOperations(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ops := result.Operations
	if len(ops) != 4 {
		t.Fatalf("expected 4 operations, got %d", len(ops))
	}

	listOp := findOp(ops, "booking.listBookingsEndpoint")
	if listOp == nil {
		t.Fatal("missing booking.listBookingsEndpoint")
	}
	if listOp.Method != "GET" {
		t.Errorf("expected GET, got %s", listOp.Method)
	}
	if listOp.Path != "/bookings" {
		t.Errorf("expected /bookings, got %s", listOp.Path)
	}
	if listOp.Summary != "List bookings" {
		t.Errorf("expected 'List bookings', got %q", listOp.Summary)
	}
	if len(listOp.QueryParams) != 2 {
		t.Errorf("expected 2 query params, got %d", len(listOp.QueryParams))
	}

	getOp := findOp(ops, "booking.getBookingEndpoint")
	if getOp == nil {
		t.Fatal("missing booking.getBookingEndpoint")
	}
	if len(getOp.PathParams) != 1 {
		t.Errorf("expected 1 path param, got %d", len(getOp.PathParams))
	}
	if getOp.PathParams[0].Name != "bookingId" {
		t.Errorf("expected bookingId path param, got %s", getOp.PathParams[0].Name)
	}
}

func TestParseCommandName(t *testing.T) {
	tests := []struct {
		operationID string
		want        string
	}{
		{"booking.listBookingsEndpoint", "list-bookings"},
		{"portfolio.getPortfolioEndpoint", "get-portfolio"},
		{"personalAccessToken.createPersonalAccessTokenEndpoint", "create-personal-access-token"},
		{"oauthAuthorize", "oauth-authorize"},
	}

	for _, tt := range tests {
		got := spec.OperationIDToCommandName(tt.operationID)
		if got != tt.want {
			t.Errorf("OperationIDToCommandName(%q) = %q, want %q", tt.operationID, got, tt.want)
		}
	}
}

func TestTagToGroupAndSubCommand(t *testing.T) {
	tests := []struct {
		tags      []string
		cmdName   string
		wantGroup string
		wantSub   string
	}{
		{[]string{"User → Bookings"}, "list-bookings", "bookings", "list"},
		{[]string{"User → Bookings"}, "get-booking", "bookings", "get"},
		{[]string{"User → Bookings"}, "create-booking", "bookings", "create"},
		{[]string{"User → Bookings"}, "delete-booking", "bookings", "delete"},
		{[]string{"Portfolio"}, "get-portfolio", "portfolio", "get"},
		{[]string{"Portfolio"}, "get-portfolio-holdings", "portfolio", "get-holdings"},
		{[]string{"Portfolio → Gains"}, "get-portfolio-gains", "gains", "get"},
		{[]string{"Portfolio → Performance"}, "get-portfolio-performance-chart", "performance", "get-chart"},
		{[]string{"Portfolio → Investment Strategies"}, "list-investment-strategies", "investment-strategies", "list"},
		{[]string{"Reference → Symbols"}, "search-symbols", "symbols", "search"},
		{[]string{"Reference → Symbols"}, "get-symbol", "symbols", "get"},
		{[]string{"User → Personal Access Tokens"}, "create-personal-access-token", "personal-access-tokens", "create"},
		{[]string{"User → OAuth Authorizations"}, "revoke-all-authorizations", "oauth-authorizations", "revoke-all"},
		{[]string{"Authentication"}, "oauth-authorize", "authentication", "oauth-authorize"},
		{[]string{}, "some-command", "", "some-command"},
	}

	for _, tt := range tests {
		group, sub := spec.TagToGroupAndSubCommand(tt.tags, tt.cmdName)
		if group != tt.wantGroup || sub != tt.wantSub {
			t.Errorf("TagToGroupAndSubCommand(%v, %q) = (%q, %q), want (%q, %q)",
				tt.tags, tt.cmdName, group, sub, tt.wantGroup, tt.wantSub)
		}
	}
}

func findOp(ops []spec.Operation, id string) *spec.Operation {
	for i := range ops {
		if ops[i].OperationID == id {
			return &ops[i]
		}
	}
	return nil
}
