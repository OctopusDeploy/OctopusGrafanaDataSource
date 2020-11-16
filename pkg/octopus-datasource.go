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

// getConnectionDetails returns the details for connecting to Octopus
func getConnectionDetails(context backend.PluginContext) (string, string) {
	var jsonData datasourceModel
	json.Unmarshal(context.DataSourceInstanceSettings.JSONData, &jsonData)
	apiKey := context.DataSourceInstanceSettings.DecryptedSecureJSONData["apiKey"]
	return jsonData.Server, apiKey
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifer).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (td *SampleDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	server, apiKey := getConnectionDetails(req.PluginContext)

	// get a mapping of space names to ids
	spaces, err := getAllResources("spaces", server, "", apiKey)
	if err != nil {
		return nil, err
	}

	// Get an array of parsed queries, with links back to the original backend query request, and maps of entities and data
	// from the Octopus REST API
	queries, data, generalEntityData, err := prepareQueries(req, server, apiKey, spaces)
	if err != nil {
		return nil, err
	}

	// Use the cache of data we returned with the call to prepareQueries() to build the grafana response
	response := td.processQueries(ctx, queries, server, apiKey, spaces, data, generalEntityData)

	return response, nil
}

// processQueries converts the data returned from the Octopus REST APIs to data to be returned to grafana
func (td *SampleDatasource) processQueries(ctx context.Context, queries []*queryModel, server string, apiKey string, spaces map[string]string, data map[string]*Deployments, generalEntityData map[string]map[string]string) (response *backend.QueryDataResponse) {
	// create response struct
	response = backend.NewQueryDataResponse()

	// We now have a list of queries, the URLs we would use to get the data, and a map of those URLs to the results
	// of the API requests. So we can no go ahead and build the response.
	for _, q := range queries {

		if q.Format == "table" {
			response.Responses[q.Query.RefID] = td.queryTable(ctx, *q, *data[q.OctopusQueryUrl])
		} else if q.Format == "timeseries" {
			response.Responses[q.Query.RefID] = td.query(ctx, *q, q.Query, *data[q.OctopusQueryUrl], server, q.SpaceName, spaces, apiKey)
		} else {
			// Any other format is the name of a resource that has an "all" endpoint in Octopus, which we retrieve as a table
			response.Responses[q.Query.RefID], _ = td.queryResources(generalEntityData[q.OctopusQueryUrl], q.Format)
		}
	}

	return response
}

// getMaps returns a map of space names to ids, maps of space names to project names to ids, and maps of space name to environment names to ids
func getMaps(req *backend.QueryDataRequest, server string, apiKey string) (spaces map[string]string, projectsMap map[string]map[string]string, environmentsMap map[string]map[string]string, err error) {
	// get a mapping of space names to ids
	spaces, err = getAllResources("spaces", server, "", apiKey)
	if err != nil {
		return nil, nil, nil, err
	}

	// get the projects and environments for the queried spaces
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

	return spaces, projectsMap, environmentsMap, nil
}

// prepareQueries looks through the queries, groups Octopus API calls to improve performance and remove redundant API calls, and returns the raw Octopus data
func prepareQueries(req *backend.QueryDataRequest, server string, apiKey string, spaces map[string]string) (queries []*queryModel, data map[string]*Deployments, generalEntityData map[string]map[string]string, err error) {
	earliestDate, latestDate := getQueryDetails(req)

	spaces, projectsMap, environmentsMap, err := getMaps(req, server, apiKey)
	if err != nil {
		return nil, nil, nil, err
	}

	// an array of parsed queries, with links back to the original backend query request
	queries = []*queryModel{}
	// A map of the Octopus query urls to the resulting deployments.
	// This map means duplicate API queries are only requested once.
	data = make(map[string]*Deployments)
	// A map of the Octopus REST API "all" endpoints we want to query.
	// Again this is used to remove duplicate API queries.
	generalEntityData = make(map[string]map[string]string)

	for i := 0; i < len(req.Queries); i++ {
		// parse the query JSON into a struct
		qm, _ := getQueryModel(req.Queries[i].JSON)
		// link back to the original backend query data
		qm.Query = req.Queries[i]
		// The list of parsed queries is a return value
		queries = append(queries, &qm)

		// get the deployments for each query
		if qm.Format == "table" || qm.Format == "timeseries" {
			// the reporting endpoint is unique in that it returns XML
			query := ""

			// Build the Octopus API URL
			if empty(qm.SpaceName) {
				query = server + "/api/reporting/deployments/xml?" +
					"fromCompletedTime=" + url.QueryEscape(earliestDate.Format(octopusDateFormat)) +
					"&toCompletedTime=" + url.QueryEscape(latestDate.Format(octopusDateFormat))
			} else {
				query = server + "/api/" + spaces[qm.SpaceName] + "/reporting/deployments/xml?" +
					"fromCompletedTime=" + url.QueryEscape(earliestDate.Format(octopusDateFormat)) +
					"&toCompletedTime=" + url.QueryEscape(latestDate.Format(octopusDateFormat))
			}

			// Filter server side on the project
			if val, ok := projectsMap[qm.SpaceName][qm.ProjectName]; ok && !empty(qm.ProjectName) {
				query += "&projectId=" + url.QueryEscape(val)
			}

			// Filter server side on the environment
			if val, ok := environmentsMap[qm.SpaceName][qm.EnvironmentName]; ok && !empty(qm.EnvironmentName) {
				query += "&environmentId=" + url.QueryEscape(val)
			}

			// Each query tracks the url that would generate the data.
			qm.OctopusQueryUrl = query

			// If the query url has not been accessed, hit the API to get the deployments.
			if _, ok := data[query]; !ok {
				xmlData, err := createRequest(query, apiKey)
				if err == nil {
					// populate the data map with the results of the API query
					data[query] = &Deployments{}
					xml.Unmarshal(xmlData, data[query])
				}
			}
		} else {
			// General entity endpoints return JSON, and can be retrieved via getAllResources()
			url := getResourceUrl(qm.Format, server, qm.SpaceName)
			// Each query tracks the url that would generate the data.
			qm.OctopusQueryUrl = url
			// Get the entities if we haven't looked them up already
			if _, ok := generalEntityData[url]; !ok {
				entities, _ := getAllResources(qm.Format, server, qm.SpaceName, apiKey)
				// populate the generalEntityData map with the results of the API query
				generalEntityData[url] = entities
			}
		}
	}

	return queries, data, generalEntityData, nil
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
