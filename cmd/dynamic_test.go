package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/PortfolioDividendTracker/cli/cmd"
)

func TestRegisterDynamicCommands(t *testing.T) {
	specJSON := `{
		"openapi": "3.1.0",
		"info": {"title": "Test", "version": "1.0"},
		"paths": {
			"/bookings": {
				"get": {
					"operationId": "booking.listBookingsEndpoint",
					"summary": "List bookings",
					"parameters": [
						{"name": "page", "in": "query", "schema": {"type": "integer"}, "description": "Page number"}
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
				}
			}
		}
	}`

	path := filepath.Join(t.TempDir(), "openapi.json")
	os.WriteFile(path, []byte(specJSON), 0600)

	root := cmd.NewRootCmd()
	err := cmd.RegisterDynamicCommands(root, path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	commands := root.Commands()
	names := make(map[string]bool)
	for _, c := range commands {
		names[c.Name()] = true
	}

	if !names["list-bookings"] {
		t.Error("missing list-bookings command")
	}
	if !names["get-booking"] {
		t.Error("missing get-booking command")
	}

	for _, c := range commands {
		if c.Name() == "list-bookings" {
			if c.Short != "List bookings" {
				t.Errorf("expected 'List bookings' short, got %q", c.Short)
			}
			f := c.Flags().Lookup("page")
			if f == nil {
				t.Error("missing --page flag on list-bookings")
			}
		}
		if c.Name() == "get-booking" {
			f := c.Flags().Lookup("bookingId")
			if f == nil {
				t.Error("missing --bookingId flag on get-booking")
			}
		}
	}
}
