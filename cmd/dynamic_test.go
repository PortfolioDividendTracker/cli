package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/PortfolioDividendTracker/cli/cmd"
	"github.com/spf13/cobra"
)

func TestRegisterDynamicCommandsGrouped(t *testing.T) {
	specJSON := `{
		"openapi": "3.1.0",
		"info": {"title": "Test", "version": "1.0"},
		"paths": {
			"/bookings": {
				"get": {
					"operationId": "booking.listBookingsEndpoint",
					"summary": "List bookings",
					"tags": ["User → Bookings"],
					"parameters": [
						{"name": "page", "in": "query", "schema": {"type": "integer"}, "description": "Page number"}
					]
				}
			},
			"/bookings/{bookingId}": {
				"get": {
					"operationId": "booking.getBookingEndpoint",
					"summary": "Get a booking",
					"tags": ["User → Bookings"],
					"parameters": [
						{"name": "bookingId", "in": "path", "required": true, "schema": {"type": "integer"}, "description": "Booking ID"}
					]
				},
				"delete": {
					"operationId": "booking.deleteBookingEndpoint",
					"summary": "Delete a booking",
					"tags": ["User → Bookings"],
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

	// Root should have a "bookings" group command
	var bookingsCmd *cobra.Command
	for _, c := range root.Commands() {
		if c.Name() == "bookings" {
			bookingsCmd = c
			break
		}
	}
	if bookingsCmd == nil {
		t.Fatal("missing 'bookings' group command")
	}

	// Check subcommands
	subNames := make(map[string]bool)
	for _, c := range bookingsCmd.Commands() {
		subNames[c.Name()] = true
	}

	if !subNames["list"] {
		t.Error("missing 'list' subcommand under bookings")
	}
	if !subNames["get"] {
		t.Error("missing 'get' subcommand under bookings")
	}
	if !subNames["delete"] {
		t.Error("missing 'delete' subcommand under bookings")
	}

	// Check flags on subcommands
	for _, c := range bookingsCmd.Commands() {
		if c.Name() == "list" {
			f := c.Flags().Lookup("page")
			if f == nil {
				t.Error("missing --page flag on bookings list")
			}
		}
		if c.Name() == "get" {
			f := c.Flags().Lookup("bookingId")
			if f == nil {
				t.Error("missing --bookingId flag on bookings get")
			}
		}
	}
}
