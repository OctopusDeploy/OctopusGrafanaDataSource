package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/httpclient"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/resource/httpadapter"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// requestTimeout bounds a single request to the Octopus server. The reporting
// feed can be slow to generate for large windows.
const requestTimeout = 100 * time.Second

var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ backend.CallResourceHandler   = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

// Datasource is created once per configured datasource instance and disposed
// when the configuration changes. All state, including caches, is scoped to
// the instance so nothing is ever shared between servers or credentials.
type Datasource struct {
	settings        Settings
	client          *octopusClient
	resourceHandler backend.CallResourceHandler
}

// NewDatasource creates a new datasource instance.
func NewDatasource(ctx context.Context, source backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	settings, err := LoadSettings(source)
	if err != nil {
		return nil, err
	}

	httpOptions, err := source.HTTPClientOptions(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not read HTTP client options: %w", err)
	}
	httpOptions.Timeouts.Timeout = requestTimeout

	httpClient, err := httpclient.New(httpOptions)
	if err != nil {
		return nil, fmt.Errorf("could not create the HTTP client: %w", err)
	}

	ds := &Datasource{
		settings: settings,
		client: &octopusClient{
			httpClient:    httpClient,
			server:        settings.Server,
			apiKey:        settings.APIKey,
			cacheDuration: settings.CacheDuration,
			cache:         newTTLCache(),
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /spaces/nameid", ds.handleSpaces)
	mux.HandleFunc("GET /{spaceId}/nameid/{resourceType}", ds.handleSpaceResources)
	ds.resourceHandler = httpadapter.New(mux)

	return ds, nil
}

// Dispose cleans up before a new instance is created for updated settings.
func (d *Datasource) Dispose() {
	d.client.httpClient.CloseIdleConnections()
}

// CallResource serves the endpoints used by the query and variable editors.
func (d *Datasource) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	return d.resourceHandler.CallResource(ctx, req, sender)
}

// handleSpaces returns a map of space names to space IDs.
func (d *Datasource) handleSpaces(rw http.ResponseWriter, req *http.Request) {
	spaces, err := d.client.getSpaces(req.Context())
	if err != nil {
		writeJSONError(rw, http.StatusBadGateway, "could not list spaces")
		return
	}
	writeJSON(rw, spaces)
}

// handleSpaceResources returns a map of entity names to IDs for one resource
// type within a space.
func (d *Datasource) handleSpaceResources(rw http.ResponseWriter, req *http.Request) {
	spaceID := req.PathValue("spaceId")
	resourceType := req.PathValue("resourceType")

	if err := validateResourceID(spaceID); err != nil {
		writeJSONError(rw, http.StatusBadRequest, "invalid space ID")
		return
	}
	if err := validateResourceType(resourceType); err != nil {
		writeJSONError(rw, http.StatusBadRequest, "unsupported resource type")
		return
	}

	entities, err := d.client.getAllResources(req.Context(), resourceType, spaceID)
	if err != nil {
		writeJSONError(rw, http.StatusBadGateway, "could not list "+resourceType)
		return
	}
	writeJSON(rw, entities)
}

func writeJSON(rw http.ResponseWriter, value any) {
	rw.Header().Set("Content-Type", "application/json")
	body, err := json.Marshal(value)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, _ = rw.Write(body)
}

func writeJSONError(rw http.ResponseWriter, status int, message string) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)
	body, _ := json.Marshal(map[string]string{"error": message})
	_, _ = rw.Write(body)
}

// QueryData handles the queries of every panel that uses this datasource.
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	response := backend.NewQueryDataResponse()

	spaces, err := d.client.getSpaces(ctx)
	if err != nil {
		for _, q := range req.Queries {
			response.Responses[q.RefID] = backend.ErrDataResponse(backend.StatusInternal, "could not list Octopus spaces: "+err.Error())
		}
		return response, nil
	}

	// The reporting feed is queried once for the widest requested window and
	// shared between the queries in this request.
	earliestDate, latestDate := getQueryDetails(req)

	state := &requestState{
		spaces:              spaces,
		earliestDate:        earliestDate,
		latestDate:          latestDate,
		reportingData:       map[string]*Deployments{},
		entityData:          map[string]map[string]string{},
		projectsBySpace:     map[string]map[string]string{},
		environmentsBySpace: map[string]map[string]string{},
	}

	for _, query := range req.Queries {
		response.Responses[query.RefID] = d.handleQuery(ctx, query, state)
	}

	return response, nil
}

// requestState caches lookups shared between the queries of one request, so
// duplicate Octopus API calls are only made once.
type requestState struct {
	spaces              map[string]string
	earliestDate        time.Time
	latestDate          time.Time
	reportingData       map[string]*Deployments
	entityData          map[string]map[string]string
	projectsBySpace     map[string]map[string]string
	environmentsBySpace map[string]map[string]string
}

func (d *Datasource) handleQuery(ctx context.Context, query backend.DataQuery, state *requestState) backend.DataResponse {
	var qm queryModel
	if err := json.Unmarshal(query.JSON, &qm); err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, "could not parse the query: "+err.Error())
	}
	qm.Query = query

	spaceID := ""
	if !empty(qm.SpaceName) {
		spaceID = state.spaces[qm.SpaceName]
		if spaceID == "" {
			return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("space %q was not found", qm.SpaceName))
		}
	}

	switch {
	case qm.usesReportingFeed():
		return d.handleReportingQuery(ctx, qm, spaceID, state)
	case qm.Format == formatAnnotationDeployments:
		return d.handleDeploymentsAnnotationQuery(ctx, qm, spaceID, state)
	default:
		return d.handleResourceQuery(ctx, qm, spaceID, state)
	}
}

func (d *Datasource) handleReportingQuery(ctx context.Context, qm queryModel, spaceID string, state *requestState) backend.DataResponse {
	projectID, environmentID, err := d.lookupProjectAndEnvironment(ctx, qm, spaceID, state)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusInternal, err.Error())
	}

	requestURL, err := d.client.reportingURL(spaceID, projectID, environmentID, state.earliestDate, state.latestDate)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, err.Error())
	}

	deployments, ok := state.reportingData[requestURL]
	if !ok {
		deployments, err = d.client.getReportingDeployments(ctx, requestURL)
		if err != nil {
			return backend.ErrDataResponse(backend.StatusInternal, "could not query the reporting feed: "+err.Error())
		}
		state.reportingData[requestURL] = deployments
	}

	response := d.dispatchReportingQuery(ctx, qm, spaceID, deployments)
	if (qm.Format == formatTimeSeries || qm.Format == formatTable) && !anyDeploymentMatches(&qm, deployments) {
		d.appendEmptyResultNotice(ctx, &response, spaceID, projectID, environmentID)
	}
	return response
}

func (d *Datasource) dispatchReportingQuery(ctx context.Context, qm queryModel, spaceID string, deployments *Deployments) backend.DataResponse {
	switch qm.Format {
	case formatTimeSeries:
		return d.queryTimeSeries(ctx, qm, spaceID, *deployments)
	case formatAnnotationReport:
		return d.queryAnnotationReport(qm, *deployments)
	default:
		return d.queryTable(qm, *deployments)
	}
}

// anyDeploymentMatches reports whether any deployment from the reporting feed
// satisfies the query filters.
func anyDeploymentMatches(qm *queryModel, deployments *Deployments) bool {
	for i := range deployments.Deployments {
		if includeDeployment(qm, &deployments.Deployments[i]) {
			return true
		}
	}
	return false
}

// appendEmptyResultNotice explains an empty result on the panel, as an empty
// panel is usually a time range that predates or postdates the deployment
// history rather than an error. The newest matching deployment is looked up
// so the user knows what range would return data.
func (d *Datasource) appendEmptyResultNotice(ctx context.Context, response *backend.DataResponse, spaceID string, projectID string, environmentID string) {
	text := "No deployments completed in the selected time range for the selected filters."
	if latest, ok := d.latestDeploymentTime(ctx, spaceID, projectID, environmentID); ok {
		text = fmt.Sprintf(
			"No deployments completed in the selected time range for the selected filters. The most recent matching deployment was created %s.",
			latest.UTC().Format("2006-01-02 15:04 UTC"))
	}

	for _, frame := range response.Frames {
		frame.AppendNotices(data.Notice{Severity: data.NoticeSeverityInfo, Text: text})
	}
}

// latestDeploymentTime returns the creation time of the newest deployment
// matching the space, project and environment filters.
func (d *Datasource) latestDeploymentTime(ctx context.Context, spaceID string, projectID string, environmentID string) (time.Time, bool) {
	deployments, err := d.client.getDeployments(ctx, spaceID, projectID, environmentID)
	if err != nil {
		return time.Time{}, false
	}

	latest := time.Time{}
	for _, deployment := range deployments {
		if deployment.CreatedParsed.After(latest) {
			latest = deployment.CreatedParsed
		}
	}
	return latest, !latest.IsZero()
}

func (d *Datasource) handleDeploymentsAnnotationQuery(ctx context.Context, qm queryModel, spaceID string, state *requestState) backend.DataResponse {
	projectID, environmentID, err := d.lookupProjectAndEnvironment(ctx, qm, spaceID, state)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusInternal, err.Error())
	}

	deployments, err := d.client.getDeployments(ctx, spaceID, projectID, environmentID)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusInternal, "could not query deployments: "+err.Error())
	}

	return queryAnnotationDeployments(deployments)
}

func (d *Datasource) handleResourceQuery(ctx context.Context, qm queryModel, spaceID string, state *requestState) backend.DataResponse {
	if err := validateResourceType(qm.Format); err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("unsupported query format %q", qm.Format))
	}

	cacheKey := spaceID + "/" + qm.Format
	entities, ok := state.entityData[cacheKey]
	if !ok {
		var err error
		entities, err = d.client.getAllResources(ctx, qm.Format, spaceID)
		if err != nil {
			return backend.ErrDataResponse(backend.StatusInternal, fmt.Sprintf("could not list %s: %s", qm.Format, err.Error()))
		}
		state.entityData[cacheKey] = entities
	}

	return queryResources(entities, qm.Format)
}

// lookupProjectAndEnvironment resolves the project and environment names of a
// query to IDs, caching the lookups per space for the current request. A name
// that cannot be resolved falls back to an empty ID, matching the original
// behaviour of filtering nothing rather than failing the query.
func (d *Datasource) lookupProjectAndEnvironment(ctx context.Context, qm queryModel, spaceID string, state *requestState) (string, string, error) {
	projectID := ""
	environmentID := ""

	if !empty(qm.ProjectName) {
		projects, ok := state.projectsBySpace[spaceID]
		if !ok {
			var err error
			projects, err = d.client.getAllResources(ctx, "projects", spaceID)
			if err != nil {
				return "", "", fmt.Errorf("could not list projects: %w", err)
			}
			state.projectsBySpace[spaceID] = projects
		}
		projectID = projects[qm.ProjectName]
	}

	if !empty(qm.EnvironmentName) {
		environments, ok := state.environmentsBySpace[spaceID]
		if !ok {
			var err error
			environments, err = d.client.getAllResources(ctx, "environments", spaceID)
			if err != nil {
				return "", "", fmt.Errorf("could not list environments: %w", err)
			}
			state.environmentsBySpace[spaceID] = environments
		}
		environmentID = environments[qm.EnvironmentName]
	}

	return projectID, environmentID, nil
}

// getQueryDetails returns the widest time range across all queries.
func getQueryDetails(req *backend.QueryDataRequest) (time.Time, time.Time) {
	earliestDate := time.Time{}
	latestDate := time.Time{}

	for i := 0; i < len(req.Queries); i++ {
		if earliestDate.IsZero() || req.Queries[i].TimeRange.From.Before(earliestDate) {
			earliestDate = req.Queries[i].TimeRange.From
		}

		if latestDate.IsZero() || req.Queries[i].TimeRange.To.After(latestDate) {
			latestDate = req.Queries[i].TimeRange.To
		}
	}

	return earliestDate, latestDate
}

// CheckHealth implements the "Save & test" button on the configuration page.
func (d *Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	apiURL, err := url.JoinPath(d.settings.Server, "api")
	if err != nil {
		return healthError("The Octopus server URL is not valid."), nil
	}

	if _, err := d.client.get(ctx, apiURL, 0); err != nil {
		return healthError("Could not contact the Octopus server. Check the server URL."), nil
	}

	meURL, err := url.JoinPath(d.settings.Server, "api", "users", "me")
	if err != nil {
		return healthError("The Octopus server URL is not valid."), nil
	}

	if _, err := d.client.get(ctx, meURL, 0); err != nil {
		return healthError("Contacted the Octopus server, but the API key was rejected."), nil
	}

	message := "Data source is working"
	if strings.HasPrefix(d.settings.Server, "http://") {
		message += ". Warning: the server URL uses plain HTTP, so the API key is sent unencrypted."
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: message,
	}, nil
}

func healthError(message string) *backend.CheckHealthResult {
	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusError,
		Message: message,
	}
}
