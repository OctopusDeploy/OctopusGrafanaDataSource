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
	"strconv"
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

func getConnectionDetails(context backend.PluginContext) (string, string, string, time.Duration) {
	var jsonData datasourceModel
	json.Unmarshal(context.DataSourceInstanceSettings.JSONData, &jsonData)
	apiKey := context.DataSourceInstanceSettings.DecryptedSecureJSONData["apiKey"]

	duration, err := strconv.Atoi(jsonData.BucketDuration)
	if err != nil {
		duration = 60
	}

	return jsonData.Server, jsonData.SpaceId, apiKey, time.Duration(duration)
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifer).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (td *SampleDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	// create response struct
	response := backend.NewQueryDataResponse()

	server, space, apiKey, bucketDuration := getConnectionDetails(req.PluginContext)

	earliestDate, latestDate, project, environment := getQueryDetails(req, server, space, apiKey)

	for i := 0; i < len(req.Queries); i++ {
		if earliestDate.Equal(time.Time{}) || req.Queries[i].TimeRange.From.Before(earliestDate) {
			earliestDate = req.Queries[i].TimeRange.From
		}

		if latestDate.Equal(time.Time{}) || req.Queries[i].TimeRange.To.After(latestDate) {
			latestDate = req.Queries[i].TimeRange.To
		}
	}

	query := server + "/api/" + space + "/reporting/deployments/xml?apikey=" + apiKey +
		"&fromCompletedTime=" + url.QueryEscape(earliestDate.Format(octopusDateFormat)) +
		"&toCompletedTime=" + url.QueryEscape(latestDate.Format(octopusDateFormat))

	if project != "" {
		query += "&projectId=" + project
	}

	if environment != "" {
		query += "&environmentId=" + environment
	}

	result, err := httpGet(query)

	if err != nil {
		return response, nil
	}
	parsedResults := Deployments{}
	xml.Unmarshal(result, &parsedResults)

	log.DefaultLogger.Info("Octopus result count " + strconv.Itoa(len(parsedResults.Deployments)))

	// loop over queries and execute them individually.
	for _, q := range req.Queries {

		var qm queryModel

		dataResponse := backend.DataResponse{}

		dataResponse.Error = json.Unmarshal(q.JSON, &qm)
		if dataResponse.Error == nil && qm.Format == "table" {
			response.Responses[q.RefID] = td.queryTable(ctx, q, parsedResults, bucketDuration)
		} else {
			response.Responses[q.RefID] = td.query(ctx, q, parsedResults, bucketDuration)
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

	path, space, apiKey, _ := getConnectionDetails(req.PluginContext)

	_, err := httpGet(path + "/api/" + space + "/reporting/deployments/xml?apikey=" + apiKey)

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
