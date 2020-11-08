package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"io/ioutil"
	"net/http"
	"time"
)

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

type jsonData struct {
	Path string
}

func getConnectionDetails(context backend.PluginContext) (string, string) {
	var jsonData jsonData
	json.Unmarshal(context.DataSourceInstanceSettings.JSONData, &jsonData)
	apiKey := context.DataSourceInstanceSettings.DecryptedSecureJSONData["apiKey"]

	return jsonData.Path, apiKey
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifer).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (td *SampleDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	log.DefaultLogger.Info("QueryData", "request", req)

	// create response struct
	response := backend.NewQueryDataResponse()

	path, apiKey := getConnectionDetails(req.PluginContext)

	result, err := httpGet(path + "/api/reporting/deployments/xml?apikey=" + apiKey)
	if err != nil {
		return response, nil
	}
	parsedResults := Deployments{}
	xml.Unmarshal(result, &parsedResults)

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := td.query(ctx, q, parsedResults)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

type queryModel struct {
	Format string `json:"format"`
}

func httpGet(url string) (result []byte, err error) {
	resp, err := http.Get(url)
	defer resp.Body.Close()

	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func extractStringColumn(deployments Deployments, f func(deployment Deployment) string) []string {
	var column []string
	for _, deployment := range deployments.Deployments {
		column = append(column, f(deployment)) // note the = instead of :=
	}
	return column
}

func extractIntColumn(deployments Deployments, f func(deployment Deployment) uint8) []uint8 {
	var column []uint8
	for _, deployment := range deployments.Deployments {
		column = append(column, f(deployment)) // note the = instead of :=
	}
	return column
}

func extractTimeColumn(deployments Deployments, f func(deployment Deployment) string) []time.Time {
	var column []time.Time
	for _, deployment := range deployments.Deployments {
		parsedTime, err := time.Parse("2006-01-02T15:04:05", f(deployment))
		if err == nil {
			column = append(column, parsedTime) // note the = instead of :=
		} else {
			column = append(column, time.Now())
		}
	}
	return column
}

func (td *SampleDatasource) query(ctx context.Context, query backend.DataQuery, deployments Deployments) backend.DataResponse {
	// Unmarshal the json into our queryModel
	var qm queryModel

	response := backend.DataResponse{}

	response.Error = json.Unmarshal(query.JSON, &qm)
	if response.Error != nil {
		return response
	}

	// create data frame response
	frame := data.NewFrame("response")

	// add the cloumns
	frame.Fields = append(frame.Fields,
		data.NewField("deploymentid", nil, extractStringColumn(deployments, func(deployment Deployment) string { return deployment.DeploymentId })),
	)

	frame.Fields = append(frame.Fields,
		data.NewField("deploymentname", nil, extractStringColumn(deployments, func(deployment Deployment) string { return deployment.DeploymentName })),
	)

	frame.Fields = append(frame.Fields,
		data.NewField("projectid", nil, extractStringColumn(deployments, func(deployment Deployment) string { return deployment.ProjectId })),
	)

	frame.Fields = append(frame.Fields,
		data.NewField("projectname", nil, extractStringColumn(deployments, func(deployment Deployment) string { return deployment.ProjectName })),
	)

	frame.Fields = append(frame.Fields,
		data.NewField("projectslug", nil, extractStringColumn(deployments, func(deployment Deployment) string { return deployment.ProjectSlug })),
	)

	frame.Fields = append(frame.Fields,
		data.NewField("tenantid", nil, extractStringColumn(deployments, func(deployment Deployment) string { return deployment.TenantId })),
	)

	frame.Fields = append(frame.Fields,
		data.NewField("tenantname", nil, extractStringColumn(deployments, func(deployment Deployment) string { return deployment.TenantName })),
	)

	frame.Fields = append(frame.Fields,
		data.NewField("channelid", nil, extractStringColumn(deployments, func(deployment Deployment) string { return deployment.ChannelId })),
	)

	frame.Fields = append(frame.Fields,
		data.NewField("channelname", nil, extractStringColumn(deployments, func(deployment Deployment) string { return deployment.ChannelName })),
	)

	frame.Fields = append(frame.Fields,
		data.NewField("environmentid", nil, extractStringColumn(deployments, func(deployment Deployment) string { return deployment.EnvironmentId })),
	)

	frame.Fields = append(frame.Fields,
		data.NewField("environmentname", nil, extractStringColumn(deployments, func(deployment Deployment) string { return deployment.EnvironmentName })),
	)

	frame.Fields = append(frame.Fields,
		data.NewField("releaseid", nil, extractStringColumn(deployments, func(deployment Deployment) string { return deployment.ReleaseId })),
	)

	frame.Fields = append(frame.Fields,
		data.NewField("releaseversion", nil, extractStringColumn(deployments, func(deployment Deployment) string { return deployment.ReleaseVersion })),
	)

	frame.Fields = append(frame.Fields,
		data.NewField("taskid", nil, extractStringColumn(deployments, func(deployment Deployment) string { return deployment.TaskId })),
	)

	frame.Fields = append(frame.Fields,
		data.NewField("taskstate", nil, extractStringColumn(deployments, func(deployment Deployment) string { return deployment.TaskState })),
	)

	frame.Fields = append(frame.Fields,
		data.NewField("deployedby", nil, extractStringColumn(deployments, func(deployment Deployment) string { return deployment.DeployedBy })),
	)

	frame.Fields = append(frame.Fields,
		data.NewField("created", nil, extractTimeColumn(deployments, func(deployment Deployment) string { return deployment.Created })),
	)

	frame.Fields = append(frame.Fields,
		data.NewField("queuetime", nil, extractTimeColumn(deployments, func(deployment Deployment) string { return deployment.QueueTime })),
	)

	frame.Fields = append(frame.Fields,
		data.NewField("starttime", nil, extractTimeColumn(deployments, func(deployment Deployment) string { return deployment.StartTime })),
	)

	frame.Fields = append(frame.Fields,
		data.NewField("competedtime", nil, extractTimeColumn(deployments, func(deployment Deployment) string { return deployment.CompletedTime })),
	)

	frame.Fields = append(frame.Fields,
		data.NewField("durationseconds", nil, extractIntColumn(deployments, func(deployment Deployment) uint8 { return deployment.DurationSeconds })),
	)

	// add the frames to the response
	response.Frames = append(response.Frames, frame)

	return response
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (td *SampleDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	log.DefaultLogger.Info("CheckHealth")

	path, apiKey := getConnectionDetails(req.PluginContext)

	_, err := httpGet(path + "/api/reporting/deployments/xml?apikey=" + apiKey)

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
	// Called before creatinga a new instance to allow plugin authors
	// to cleanup.
}
