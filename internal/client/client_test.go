package client_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PortfolioDividendTracker/cli/internal/client"
)

func TestDoGetRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer pat_test" {
			t.Errorf("expected Bearer pat_test, got %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("expected application/json accept, got %s", r.Header.Get("Accept"))
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("expected page=2, got %s", r.URL.Query().Get("page"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	c := client.New(server.URL, "pat_test")

	query := map[string]string{"page": "2"}
	result, statusCode, err := c.Do("GET", "/bookings", nil, query)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if statusCode != 200 {
		t.Errorf("expected 200, got %d", statusCode)
	}

	var body map[string]string
	if err := json.Unmarshal(result, &body); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected ok, got %s", body["status"])
	}
}

func TestDoPostRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected application/json content-type, got %s", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]int{"id": 1})
	}))
	defer server.Close()

	c := client.New(server.URL, "pat_test")

	body := []byte(`{"name": "test"}`)
	result, statusCode, err := c.Do("POST", "/bookings", body, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if statusCode != 201 {
		t.Errorf("expected 201, got %d", statusCode)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestPathParamSubstitution(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bookings/42" {
			t.Errorf("expected /bookings/42, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := client.New(server.URL, "pat_test")

	pathParams := map[string]string{"bookingId": "42"}
	_, _, err := c.DoWithPathParams("GET", "/bookings/{bookingId}", pathParams, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMissingToken(t *testing.T) {
	c := client.New("http://example.com", "")
	_, _, err := c.Do("GET", "/test", nil, nil)
	if err == nil {
		t.Fatal("expected error for missing token")
	}
}
