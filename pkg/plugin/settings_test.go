package plugin

import (
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func TestLoadSettings(t *testing.T) {
	tests := []struct {
		name       string
		jsonData   string
		wantErr    bool
		wantServer string
		wantCache  time.Duration
	}{
		{name: "valid", jsonData: `{"server":"https://example.octopus.app","cacheDuration":"1m"}`, wantServer: "https://example.octopus.app", wantCache: time.Minute},
		{name: "trailing slash trimmed", jsonData: `{"server":"https://example.octopus.app/"}`, wantServer: "https://example.octopus.app"},
		{name: "no cache duration", jsonData: `{"server":"http://octopus"}`, wantServer: "http://octopus"},
		{name: "empty server", jsonData: `{"server":""}`, wantErr: true},
		{name: "missing server", jsonData: `{}`, wantErr: true},
		{name: "bad scheme", jsonData: `{"server":"ftp://example"}`, wantErr: true},
		{name: "no host", jsonData: `{"server":"https://"}`, wantErr: true},
		{name: "bad duration", jsonData: `{"server":"https://example","cacheDuration":"soon"}`, wantErr: true},
		{name: "negative duration", jsonData: `{"server":"https://example","cacheDuration":"-1m"}`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings, err := LoadSettings(backend.DataSourceInstanceSettings{
				JSONData:                []byte(tt.jsonData),
				DecryptedSecureJSONData: map[string]string{"apiKey": "API-TEST"},
			})

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got settings %+v", settings)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if settings.Server != tt.wantServer {
				t.Errorf("server = %q, want %q", settings.Server, tt.wantServer)
			}
			if settings.CacheDuration != tt.wantCache {
				t.Errorf("cacheDuration = %v, want %v", settings.CacheDuration, tt.wantCache)
			}
			if settings.APIKey != "API-TEST" {
				t.Errorf("apiKey was not loaded from the secure settings")
			}
		})
	}
}
