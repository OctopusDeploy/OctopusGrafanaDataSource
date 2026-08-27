package plugin

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

const (
	// maxResponseBytes caps how much of an Octopus response is read. The
	// reporting feed can be large, but an unbounded read is a memory DoS risk.
	maxResponseBytes = 128 << 20

	// longCacheDuration is used for responses that never change once written,
	// such as a release or a reporting query over a fixed historical window.
	longCacheDuration = 24 * time.Hour

	// failedCacheDuration is a circuit breaker: failed requests are not
	// retried for this long (only when caching is enabled for the request).
	failedCacheDuration = time.Minute

	octopusDateFormat = "2006-01-02 15:04:05"
)

// resourceIDPattern matches Octopus resource IDs such as Spaces-1 or
// Projects-42. Everything interpolated into a request path or query string
// must match this pattern or be rejected.
var resourceIDPattern = regexp.MustCompile(`^[A-Za-z]+-[0-9]+$`)

// allowedResourceTypes is the complete set of entity collections the plugin
// will query. Requests for anything else are rejected before a URL is built.
var allowedResourceTypes = map[string]bool{
	"accounts":            true,
	"actiontemplates":     true,
	"certificates":        true,
	"channels":            true,
	"deployments":         true,
	"environments":        true,
	"feeds":               true,
	"libraryvariablesets": true,
	"machinepolicies":     true,
	"machineroles":        true,
	"machines":            true,
	"octopusservernodes":  true,
	"permissions":         true,
	"projectgroups":       true,
	"projects":            true,
	"proxies":             true,
	"releases":            true,
	"roles":               true,
	"runbooks":            true,
	"spaces":              true,
	"subscriptions":       true,
	"tagsets":             true,
	"teams":               true,
	"tenants":             true,
	"tenantvariables":     true,
	"users":               true,
	"variables":           true,
	"workerpools":         true,
	"workers":             true,
}

// spaceAgnosticResourceTypes live outside a space in the Octopus API.
var spaceAgnosticResourceTypes = map[string]bool{
	"spaces": true,
	"users":  true,
}

// resourceTypesWithoutAllEndpoint return every record from their collection
// endpoint rather than exposing an "/all" endpoint.
var resourceTypesWithoutAllEndpoint = map[string]bool{
	"deployments": true,
	"releases":    true,
}

var errInvalidResourceType = errors.New("unsupported resource type")
var errInvalidResourceID = errors.New("invalid resource ID")

// octopusClient performs all communication with a single Octopus server.
// One client exists per datasource instance, so the cache is bound to one
// (server, API key) pair and can never leak data between instances.
type octopusClient struct {
	httpClient    *http.Client
	server        string
	apiKey        string
	cacheDuration time.Duration
	cache         *ttlCache
}

func validateResourceType(resourceType string) error {
	if !allowedResourceTypes[resourceType] {
		return fmt.Errorf("%w: %q", errInvalidResourceType, resourceType)
	}
	return nil
}

func validateResourceID(id string) error {
	if !resourceIDPattern.MatchString(id) {
		return fmt.Errorf("%w: %q", errInvalidResourceID, id)
	}
	return nil
}

// validateOptionalResourceID accepts an empty string, otherwise the value
// must be a well formed resource ID.
func validateOptionalResourceID(id string) error {
	if empty(id) {
		return nil
	}
	return validateResourceID(id)
}

// get performs a GET request against the Octopus server, honouring the cache
// TTL. A TTL of zero bypasses the cache entirely.
func (c *octopusClient) get(ctx context.Context, requestURL string, ttl time.Duration) ([]byte, error) {
	if ttl > 0 {
		if value, failed, found := c.cache.get(requestURL); found {
			if failed {
				return nil, fmt.Errorf("skipping request to %s: a recent identical request failed", requestURL)
			}
			return value, nil
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Octopus-ApiKey", c.apiKey)
	req.Header.Set("Accept", "application/json, application/xml")

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to the Octopus server failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	log.DefaultLogger.Debug("Octopus request", "path", req.URL.Path, "status", resp.StatusCode, "duration", time.Since(start).String())

	if resp.StatusCode != http.StatusOK {
		if ttl > 0 {
			c.cache.set(requestURL, nil, true, failedCacheDuration)
		}
		return nil, fmt.Errorf("the Octopus server responded with status %d for %s", resp.StatusCode, req.URL.Path)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("reading the Octopus response failed: %w", err)
	}
	if len(body) > maxResponseBytes {
		return nil, fmt.Errorf("the Octopus response for %s exceeded the %d byte limit", req.URL.Path, maxResponseBytes)
	}

	if ttl > 0 {
		c.cache.set(requestURL, body, false, ttl)
	}

	return body, nil
}

// resourceURL builds the collection URL for a resource type, validating every
// component before it is interpolated.
func (c *octopusClient) resourceURL(resourceType string, spaceID string) (string, error) {
	if err := validateResourceType(resourceType); err != nil {
		return "", err
	}
	if err := validateOptionalResourceID(spaceID); err != nil {
		return "", err
	}

	segments := []string{"api"}
	if !empty(spaceID) && !spaceAgnosticResourceTypes[resourceType] {
		segments = append(segments, spaceID)
	}
	segments = append(segments, resourceType)
	if !resourceTypesWithoutAllEndpoint[resourceType] {
		segments = append(segments, "all")
	}

	return url.JoinPath(c.server, segments...)
}

// getAllResources returns a map of resource names (or versions for releases)
// to IDs for the requested resource type.
func (c *octopusClient) getAllResources(ctx context.Context, resourceType string, spaceID string) (map[string]string, error) {
	requestURL, err := c.resourceURL(resourceType, spaceID)
	if err != nil {
		return nil, err
	}

	body, err := c.get(ctx, requestURL, c.cacheDuration)
	if err != nil {
		return nil, err
	}

	var parsedResults []BaseResource
	if err := json.Unmarshal(body, &parsedResults); err != nil {
		return nil, fmt.Errorf("could not parse the %s response: %w", resourceType, err)
	}

	results := make(map[string]string)
	for _, r := range parsedResults {
		if !empty(r.Version) {
			results[r.Version] = r.Id
		} else {
			results[r.Name] = r.Id
		}
	}
	return results, nil
}

// getSpaces returns a map of space names to space IDs. The default space is
// additionally mapped from a single space character, matching how the
// frontend identifies "no space selected".
func (c *octopusClient) getSpaces(ctx context.Context) (map[string]string, error) {
	requestURL, err := c.resourceURL("spaces", "")
	if err != nil {
		return nil, err
	}

	body, err := c.get(ctx, requestURL, c.cacheDuration)
	if err != nil {
		return nil, err
	}

	var parsedResults []SpaceResource
	if err := json.Unmarshal(body, &parsedResults); err != nil {
		return nil, fmt.Errorf("could not parse the spaces response: %w", err)
	}

	results := make(map[string]string)
	for _, r := range parsedResults {
		results[r.Name] = r.Id
		if r.IsDefault {
			results[" "] = r.Id
		}
	}
	return results, nil
}

// getDeployments returns deployments from the JSON API, optionally filtered
// by project and environment.
func (c *octopusClient) getDeployments(ctx context.Context, spaceID string, projectID string, environmentID string) ([]PlainDeployment, error) {
	if err := validateOptionalResourceID(spaceID); err != nil {
		return nil, err
	}
	if err := validateOptionalResourceID(projectID); err != nil {
		return nil, err
	}
	if err := validateOptionalResourceID(environmentID); err != nil {
		return nil, err
	}

	segments := []string{"api"}
	if !empty(spaceID) {
		segments = append(segments, spaceID)
	}
	segments = append(segments, "deployments")

	requestURL, err := url.JoinPath(c.server, segments...)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	if !empty(projectID) {
		query.Set("projects", projectID)
	}
	if !empty(environmentID) {
		query.Set("environments", environmentID)
	}
	if len(query) > 0 {
		requestURL += "?" + query.Encode()
	}

	body, err := c.get(ctx, requestURL, c.cacheDuration)
	if err != nil {
		return nil, err
	}

	var parsedResults PlainDeploymentItems
	if err := json.Unmarshal(body, &parsedResults); err != nil {
		return nil, fmt.Errorf("could not parse the deployments response: %w", err)
	}

	for index := range parsedResults.Items {
		created, err := time.Parse(dateFormat, parsedResults.Items[index].Created)
		if err == nil {
			parsedResults.Items[index].CreatedParsed = created
		}
	}
	return parsedResults.Items, nil
}

// getRelease returns the details of a specific release. Releases are
// immutable, so responses are cached for a long time.
func (c *octopusClient) getRelease(ctx context.Context, spaceID string, releaseID string) (Release, error) {
	if err := validateOptionalResourceID(spaceID); err != nil {
		return Release{}, err
	}
	if err := validateResourceID(releaseID); err != nil {
		return Release{}, err
	}

	segments := []string{"api"}
	if !empty(spaceID) {
		segments = append(segments, spaceID)
	}
	segments = append(segments, "releases", releaseID)

	requestURL, err := url.JoinPath(c.server, segments...)
	if err != nil {
		return Release{}, err
	}

	body, err := c.get(ctx, requestURL, longCacheDuration)
	if err != nil {
		return Release{}, err
	}

	var parsedResults Release
	if err := json.Unmarshal(body, &parsedResults); err != nil {
		return Release{}, fmt.Errorf("could not parse the release response: %w", err)
	}

	assembled, err := time.Parse(dateFormat, parsedResults.Assembled)
	if err == nil {
		parsedResults.AssembledDate = assembled
	}
	return parsedResults, nil
}

// reportingURL builds the URL of the XML reporting feed for a completed time
// window, optionally filtered by project and environment.
func (c *octopusClient) reportingURL(spaceID string, projectID string, environmentID string, earliestDate time.Time, latestDate time.Time) (string, error) {
	if err := validateOptionalResourceID(spaceID); err != nil {
		return "", err
	}
	if err := validateOptionalResourceID(projectID); err != nil {
		return "", err
	}
	if err := validateOptionalResourceID(environmentID); err != nil {
		return "", err
	}

	segments := []string{"api"}
	if !empty(spaceID) {
		segments = append(segments, spaceID)
	}
	segments = append(segments, "reporting", "deployments", "xml")

	requestURL, err := url.JoinPath(c.server, segments...)
	if err != nil {
		return "", err
	}

	query := url.Values{}
	query.Set("fromCompletedTime", earliestDate.Format(octopusDateFormat))
	query.Set("toCompletedTime", latestDate.Format(octopusDateFormat))
	if !empty(projectID) {
		query.Set("projectId", projectID)
	}
	if !empty(environmentID) {
		query.Set("environmentId", environmentID)
	}

	return requestURL + "?" + query.Encode(), nil
}

// getReportingDeployments fetches and parses the XML reporting feed.
func (c *octopusClient) getReportingDeployments(ctx context.Context, requestURL string) (*Deployments, error) {
	// The feed only contains completed deployments for a fixed time window,
	// so it can be cached for a long time under its exact URL.
	body, err := c.get(ctx, requestURL, longCacheDuration)
	if err != nil {
		return nil, err
	}

	deployments := &Deployments{}
	if err := xml.Unmarshal(body, deployments); err != nil {
		return nil, fmt.Errorf("could not parse the reporting feed: %w", err)
	}

	parseTimes(deployments)
	return deployments, nil
}
