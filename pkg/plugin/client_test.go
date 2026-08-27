package plugin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func newTestClient(server string, cacheDuration time.Duration) *octopusClient {
	return &octopusClient{
		httpClient:    http.DefaultClient,
		server:        server,
		apiKey:        "API-TEST",
		cacheDuration: cacheDuration,
		cache:         newTTLCache(),
	}
}

func TestResourceURLValidation(t *testing.T) {
	client := newTestClient("https://example.octopus.app", 0)

	tests := []struct {
		name         string
		resourceType string
		spaceID      string
		wantURL      string
		wantErr      bool
	}{
		{name: "spaced resource", resourceType: "projects", spaceID: "Spaces-1", wantURL: "https://example.octopus.app/api/Spaces-1/projects/all"},
		{name: "space agnostic resource", resourceType: "spaces", spaceID: "Spaces-1", wantURL: "https://example.octopus.app/api/spaces/all"},
		{name: "no all endpoint", resourceType: "deployments", spaceID: "Spaces-1", wantURL: "https://example.octopus.app/api/Spaces-1/deployments?take=2147483647"},
		{name: "unspaced", resourceType: "environments", spaceID: "", wantURL: "https://example.octopus.app/api/environments/all"},
		{name: "unknown resource type", resourceType: "certificates-of-authenticity", wantErr: true},
		{name: "path traversal resource type", resourceType: "../users", wantErr: true},
		{name: "path traversal space", resourceType: "projects", spaceID: "../../api/users", wantErr: true},
		{name: "malformed space", resourceType: "projects", spaceID: "Spaces-1/extra", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := client.resourceURL(tt.resourceType, tt.spaceID)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantURL {
				t.Errorf("url = %q, want %q", got, tt.wantURL)
			}
		})
	}
}

func TestReportingURLValidation(t *testing.T) {
	client := newTestClient("https://example.octopus.app", 0)
	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC)

	got, err := client.reportingURL("Spaces-1", "Projects-2", "Environments-3", from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "https://example.octopus.app/api/Spaces-1/reporting/deployments/xml?" +
		"environmentId=Environments-3&fromCompletedTime=2020-01-01+00%3A00%3A00&projectId=Projects-2&toCompletedTime=2020-02-01+00%3A00%3A00"
	if got != want {
		t.Errorf("url = %q, want %q", got, want)
	}

	if _, err := client.reportingURL("Spaces-1", "Projects-2&admin=true", "", from, to); err == nil {
		t.Errorf("expected an error for a project ID containing query characters")
	}

	if _, err := client.reportingURL("Spaces-1", "", "../secrets", from, to); err == nil {
		t.Errorf("expected an error for an environment ID containing path characters")
	}
}

func TestGetSendsAPIKeyAndCaches(t *testing.T) {
	requests := int32(0)
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		atomic.AddInt32(&requests, 1)
		if req.Header.Get("X-Octopus-ApiKey") != "API-TEST" {
			rw.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = rw.Write([]byte(`[{"Name":"Default","Id":"Spaces-1","IsDefault":true}]`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, time.Minute)

	for i := 0; i < 3; i++ {
		spaces, err := client.getSpaces(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if spaces["Default"] != "Spaces-1" || spaces[" "] != "Spaces-1" {
			t.Fatalf("unexpected spaces map: %v", spaces)
		}
	}

	if atomic.LoadInt32(&requests) != 1 {
		t.Errorf("expected 1 upstream request thanks to caching, got %d", requests)
	}
}

func TestGetDoesNotCacheWhenDisabled(t *testing.T) {
	requests := int32(0)
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		atomic.AddInt32(&requests, 1)
		_, _ = rw.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, 0)

	for i := 0; i < 2; i++ {
		if _, err := client.getSpaces(context.Background()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if atomic.LoadInt32(&requests) != 2 {
		t.Errorf("expected 2 upstream requests with caching disabled, got %d", requests)
	}
}

func TestFailedRequestsAreCircuitBroken(t *testing.T) {
	requests := int32(0)
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		atomic.AddInt32(&requests, 1)
		rw.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := newTestClient(server.URL, time.Minute)

	for i := 0; i < 3; i++ {
		if _, err := client.getSpaces(context.Background()); err == nil {
			t.Fatalf("expected an error")
		}
	}

	if atomic.LoadInt32(&requests) != 1 {
		t.Errorf("expected 1 upstream request thanks to the circuit breaker, got %d", requests)
	}
}

// TestCacheIsScopedToInstance is a regression test for a vulnerability in the
// original plugin, where a process wide cache leaked responses between
// datasource instances with different servers or credentials.
func TestCacheIsScopedToInstance(t *testing.T) {
	makeServer := func(body string) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			_, _ = rw.Write([]byte(body))
		}))
	}

	serverA := makeServer(`[{"Name":"TeamA","Id":"Spaces-1","IsDefault":true}]`)
	defer serverA.Close()
	serverB := makeServer(`[{"Name":"TeamB","Id":"Spaces-1","IsDefault":true}]`)
	defer serverB.Close()

	clientA := newTestClient(serverA.URL, time.Hour)
	clientB := newTestClient(serverB.URL, time.Hour)

	spacesA, err := clientA.getSpaces(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	spacesB, err := clientB.getSpaces(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := spacesA["TeamA"]; !ok {
		t.Errorf("client A returned the wrong spaces: %v", spacesA)
	}
	if _, ok := spacesB["TeamB"]; !ok {
		t.Errorf("client B was served another instance's data: %v", spacesB)
	}
}

func TestGetDeploymentsEscapesQueryValues(t *testing.T) {
	var capturedQuery string
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		capturedQuery = req.URL.RawQuery
		_, _ = rw.Write([]byte(`{"Items":[]}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, 0)

	if _, err := client.getDeployments(context.Background(), "Spaces-1", "Projects-1", "Environments-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(capturedQuery, "projects=Projects-1") || !strings.Contains(capturedQuery, "environments=Environments-1") {
		t.Errorf("unexpected query string: %q", capturedQuery)
	}

	if _, err := client.getDeployments(context.Background(), "Spaces-1", "Projects-1&take=9999", ""); err == nil {
		t.Errorf("expected an error for a project ID containing query characters")
	}
}
