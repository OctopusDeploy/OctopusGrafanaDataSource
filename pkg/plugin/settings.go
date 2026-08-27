package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// Settings holds the validated datasource configuration.
type Settings struct {
	// Server is the Octopus server base URL without a trailing slash.
	Server string
	// CacheDuration is how long entity lookups (projects, environments, ...)
	// are cached. Zero disables entity caching.
	CacheDuration time.Duration
	// APIKey is the Octopus API key. It must never be logged or returned to
	// the frontend.
	APIKey string
}

type settingsJSON struct {
	Server        string `json:"server"`
	CacheDuration string `json:"cacheDuration"`
}

// LoadSettings parses and validates the datasource instance settings.
func LoadSettings(source backend.DataSourceInstanceSettings) (Settings, error) {
	parsed := settingsJSON{}
	if err := json.Unmarshal(source.JSONData, &parsed); err != nil {
		return Settings{}, fmt.Errorf("could not parse datasource settings: %w", err)
	}

	server, err := validateServerURL(parsed.Server)
	if err != nil {
		return Settings{}, err
	}

	cacheDuration := time.Duration(0)
	if strings.TrimSpace(parsed.CacheDuration) != "" {
		cacheDuration, err = time.ParseDuration(strings.TrimSpace(parsed.CacheDuration))
		if err != nil || cacheDuration < 0 {
			return Settings{}, fmt.Errorf("cache duration %q is not a valid duration (e.g. 1m, 30s)", parsed.CacheDuration)
		}
	}

	return Settings{
		Server:        server,
		CacheDuration: cacheDuration,
		APIKey:        source.DecryptedSecureJSONData["apiKey"],
	}, nil
}

func validateServerURL(server string) (string, error) {
	trimmed := strings.TrimRight(strings.TrimSpace(server), "/")
	if trimmed == "" {
		return "", errors.New("the Octopus server URL is not configured")
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("the Octopus server URL is not valid: %w", err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errors.New("the Octopus server URL must use http or https")
	}

	if parsed.Host == "" {
		return "", errors.New("the Octopus server URL must include a host")
	}

	return trimmed, nil
}
