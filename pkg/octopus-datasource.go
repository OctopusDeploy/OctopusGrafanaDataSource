package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"net/http"
	"net/url"
	"time"
)

const octopusDateFormat = "2006-01-02 15:04:05"
const maxFrames = 50

// newDatasource returns datasource.ServeOpts.
func newDatasource() datasource.ServeOpts {
	// creates a instance manager for your plugin. The function passed
	// into `NewInstanceManger` is called when the instance is created
	// for the first time or when a datasource configuration changed.
	im := datasource.NewInstanceManager(newDataSourceInstance)
	ds := &SampleDatasource{
		im: im,
	}

	return datasource.ServeOpts{
		QueryDataHandler:   ds,
		CheckHealthHandler: ds,
	}
}

// SampleDatasource is an example datasource used to scaffold
// new datasource plugins with an backend.
type SampleDatasource struct {
	// The instance manager can help with lifecycle management
	// of datasource instances in plugins. It's not a requirements
	// but a best practice that we recommend that you follow.
	im instancemgmt.InstanceManager
}

func getConnectionDetails(context backend.PluginContext) (string, string) {
	var jsonData datasourceModel
	json.Unmarshal(context.DataSourceInstanceSettings.JSONData, &jsonData)
	apiKey := context.DataSourceInstanceSettings.DecryptedSecureJSONData["apiKey"]
	return jsonData.Server, apiKey
}

func getDeploymentHistory(server string, spaceId string, apiKey string, earliestDate time.Time, latestDate time.Time) (Deployments, error) {
	query := server + "/api/" + spaceId + "/reporting/deployments/xml?apikey=" + apiKey +
		"&fromCompletedTime=" + url.QueryEscape(earliestDate.Format(octopusDateFormat)) +
		"&toCompletedTime=" + url.QueryEscape(latestDate.Format(octopusDateFormat))

	result, err := createRequest(query, apiKey)
	parsedResults := Deployments{}

	if err != nil {
		return parsedResults, nil
	}

	xml.Unmarshal(result, &parsedResults)
	return parsedResults, nil
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifer).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (td *SampleDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	// create response struct
	response := backend.NewQueryDataResponse()

	server, apiKey := getConnectionDetails(req.PluginContext)

	earliestDate, latestDate := getQueryDetails(req)

	for i := 0; i < len(req.Queries); i++ {
		if earliestDate.Equal(time.Time{}) || req.Queries[i].TimeRange.From.Before(earliestDate) {
			earliestDate = req.Queries[i].TimeRange.From
		}

		if latestDate.Equal(time.Time{}) || req.Queries[i].TimeRange.To.After(latestDate) {
			latestDate = req.Queries[i].TimeRange.To
		}
	}

	// get a mapping of space names to ids
	spaces, err := getAllResources("spaces", server, "", apiKey)

	// get the projects and environments for the queried spaces
	projectsMap := map[string]map[string]string{}
	environmentsMap := map[string]map[string]string{}
	for i := 0; i < len(req.Queries); i++ {
		qm, _ := getQueryModel(req.Queries[i].JSON)

		if _, ok := projectsMap[qm.SpaceName]; !ok {
			projects, _ := getAllResources("projects", server, spaces[qm.SpaceName], apiKey)
			projectsMap[qm.SpaceName] = projects
		}

		if _, ok := environmentsMap[qm.SpaceName]; !ok {
			environments, _ := getAllResources("environments", server, spaces[qm.SpaceName], apiKey)
			environmentsMap[qm.SpaceName] = environments
		}
	}

	if err != nil {
		return nil, err
	}

	// an array of parsed queries, with links back to the original backend query request
	queries := []*queryModel{}
	// A map of the Octopus query urls to the resulting deployments.
	// This map means duplicate queries are only requested once.
	data := make(map[string]*Deployments)

	for i := 0; i < len(req.Queries); i++ {
		// parse the query JSON into a struct
		qm, _ := getQueryModel(req.Queries[i].JSON)
		// link back to the original backend query data
		qm.Query = req.Queries[i]
		// we'll loop over these queries later
		queries = append(queries, &qm)

		// get the deployments for each query
		if qm.Format == "table" || qm.Format == "timeseries" {

			query := ""

			if empty(qm.SpaceName) {
				query = server + "/api/reporting/deployments/xml?" +
					"fromCompletedTime=" + url.QueryEscape(earliestDate.Format(octopusDateFormat)) +
					"&toCompletedTime=" + url.QueryEscape(latestDate.Format(octopusDateFormat))
			} else {
				query = server + "/api/" + spaces[qm.SpaceName] + "/reporting/deployments/xml?" +
					"fromCompletedTime=" + url.QueryEscape(earliestDate.Format(octopusDateFormat)) +
					"&toCompletedTime=" + url.QueryEscape(latestDate.Format(octopusDateFormat))
			}

			if val, ok := projectsMap[qm.SpaceName][qm.ProjectName]; ok && !empty(qm.ProjectName) {
				query += "&projectId=" + url.QueryEscape(val)
			}

			if val, ok := environmentsMap[qm.SpaceName][qm.EnvironmentName]; ok && !empty(qm.EnvironmentName) {
				query += "&environmentId=" + url.QueryEscape(val)
			}

			// Each query tracks the url that would generate the data.
			qm.OctopusQueryUrl = query

			// If the query url has not been accessed, hit the API to get the deployments.
			if _, ok := data[query]; !ok {
				xmlData, err := createRequest(query, apiKey)
				if err == nil {
					data[query] = &Deployments{}
					xml.Unmarshal(xmlData, data[query])
				}
			}
		}
	}

	// loop over queries and execute them individually.
	for _, q := range queries {

		if q.Format == "table" {
			response.Responses[q.Query.RefID] = td.queryTable(ctx, *q, *data[q.OctopusQueryUrl])
		} else if q.Format == "timeseries" {
			response.Responses[q.Query.RefID] = td.query(ctx, *q, q.Query, *data[q.OctopusQueryUrl], server, q.SpaceName, spaces, apiKey)
		} else {
			// Any other format is the name of a resource that has an "all" endpoint in Octopus, which we retrieve as a table
			response.Responses[q.Query.RefID], _ = td.queryResources(q.Format, spaces[q.SpaceName], ctx, req)
		}
	}

	return response, nil
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (td *SampleDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	log.DefaultLogger.Info("CheckHealth")

	path, apiKey := getConnectionDetails(req.PluginContext)

	_, err := createRequest(path+"/api", apiKey)

	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "failed to contact Octopus server",
		}, nil
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Data source is working",
	}, nil
}

type instanceSettings struct {
	httpClient *http.Client
}

func newDataSourceInstance(setting backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	return &instanceSettings{
		httpClient: &http.Client{},
	}, nil
}

func (s *instanceSettings) Dispose() {
	// Called before creating a new instance to allow plugin authors
	// to cleanup.
}
