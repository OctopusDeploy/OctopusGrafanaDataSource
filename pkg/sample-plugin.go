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
	"strconv"
	"time"
)

const dateFormat = "2006-01-02T15:04:05"
const maxFrames = 50

func Min(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

func MinInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func Exists(arr []time.Time, item time.Time, endCheck int) bool {
	for i := 0; i < MinInt(endCheck, len(arr)); i++ {
		if arr[i] == item {
			return true
		}
	}

	return false
}

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
	Server         string
	SpaceId        string
	BucketDuration string
}

func getConnectionDetails(context backend.PluginContext) (string, string, string, time.Duration) {
	var jsonData jsonData
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
	log.DefaultLogger.Info("QueryData", "request", req)

	// create response struct
	response := backend.NewQueryDataResponse()

	server, space, apiKey, bucketDuration := getConnectionDetails(req.PluginContext)

	result, err := httpGet(server + "/api/" + space + "/reporting/deployments/xml?apikey=" + apiKey)
	if err != nil {
		return response, nil
	}
	parsedResults := Deployments{}
	xml.Unmarshal(result, &parsedResults)

	log.DefaultLogger.Info("Octopus result count " + strconv.Itoa(len(parsedResults.Deployments)))

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := td.query(ctx, q, parsedResults, bucketDuration)

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

func (td *SampleDatasource) query(ctx context.Context, query backend.DataQuery, deployments Deployments, bucketDuration time.Duration) backend.DataResponse {
	// Unmarshal the json into our queryModel
	var qm queryModel

	response := backend.DataResponse{}

	response.Error = json.Unmarshal(query.JSON, &qm)
	if response.Error != nil {
		return response
	}

	// create data frame response
	frame := data.NewFrame("response")

	// The field data
	times := []time.Time{}
	deploymentId := []string{}
	deploymentName := []string{}
	projectId := []string{}
	projectName := []string{}
	projectSlug := []string{}
	tenantId := []string{}
	tenantName := []string{}
	channelId := []string{}
	channelName := []string{}
	environmentId := []string{}
	environmentName := []string{}
	releaseId := []string{}
	releaseVersion := []string{}
	taskId := []string{}
	taskState := []string{}
	deployedBy := []string{}
	created := []string{}
	queueTime := []string{}
	startTime := []string{}
	duration := []uint8{}

	// Work out how long the buckets should be
	buckets := Min(maxFrames, int64(query.TimeRange.Duration()/bucketDuration))
	bucketDuration = query.TimeRange.Duration() / time.Duration(buckets)

	// get the bucket start time for each deployment
	for i := 0; i < len(deployments.Deployments); i++ {
		time, err := time.Parse(dateFormat, deployments.Deployments[i].CompletedTime)
		if err == nil {
			deployments.Deployments[i].CompetedTimeRounded = time.Round(bucketDuration)
		}
	}

	for i := 0; i < int(buckets); i++ {
		roundedTime := query.TimeRange.From.Add(bucketDuration * time.Duration(i)).Round(bucketDuration)
		if query.TimeRange.From.Before(roundedTime) && query.TimeRange.To.After(roundedTime) {

			found := false

			for _, d := range deployments.Deployments {

				if d.CompetedTimeRounded.Equal(roundedTime) {
					found = true
					times = append(times, roundedTime)
					deploymentId = append(deploymentId, d.DeploymentId)
					deploymentName = append(deploymentName, d.DeploymentName)
					projectId = append(projectId, d.ProjectId)
					projectName = append(projectName, d.ProjectName)
					projectSlug = append(projectSlug, d.ProjectSlug)
					tenantId = append(tenantId, d.TenantId)
					tenantName = append(tenantName, d.TenantName)
					channelId = append(channelId, d.ChannelId)
					channelName = append(channelName, d.ChannelName)
					environmentId = append(environmentId, d.EnvironmentId)
					environmentName = append(environmentName, d.EnvironmentName)
					releaseId = append(releaseId, d.ReleaseId)
					releaseVersion = append(releaseVersion, d.ReleaseVersion)
					taskId = append(taskId, d.TaskId)
					taskState = append(taskState, d.TaskState)
					deployedBy = append(deployedBy, d.DeployedBy)
					created = append(created, d.Created)
					queueTime = append(queueTime, d.QueueTime)
					startTime = append(startTime, d.StartTime)
					duration = append(duration, d.DurationSeconds)
				}
			}

			if !found {
				times = append(times, roundedTime)
				deploymentId = append(deploymentId, "")
				deploymentName = append(deploymentName, "")
				projectId = append(projectId, "")
				projectName = append(projectName, "")
				projectSlug = append(projectSlug, "")
				tenantId = append(tenantId, "")
				tenantName = append(tenantName, "")
				channelId = append(channelId, "")
				channelName = append(channelName, "")
				environmentId = append(environmentId, "")
				environmentName = append(environmentName, "")
				releaseId = append(releaseId, "")
				releaseVersion = append(releaseVersion, "")
				taskId = append(taskId, "")
				taskState = append(taskState, "")
				deployedBy = append(deployedBy, "")
				created = append(created, "")
				queueTime = append(queueTime, "")
				startTime = append(startTime, "")
				duration = append(duration, 0)
			}
		}
	}

	frame.Fields = append(frame.Fields,
		data.NewField("time", nil, times),
		/*data.NewField("deploymentid", nil, deploymentId),
		  data.NewField("deploymentname", nil, deploymentName),
		  data.NewField("projectid", nil, projectId),
		  data.NewField("projectname", nil, projectName),
		  data.NewField("projectslug", nil, projectSlug),
		  data.NewField("tenantid", nil, tenantId),
		  data.NewField("tenantname", nil, tenantName),
		  data.NewField("channelid", nil, channelId),
		  data.NewField("channelname", nil, channelName),
		  data.NewField("environmentid", nil, environmentId),
		  data.NewField("environmentname", nil, environmentName),
		  data.NewField("releaseid", nil, releaseId),
		  data.NewField("releaseversion", nil, releaseVersion),
		  data.NewField("taskid", nil, taskId),
		  data.NewField("taskstate", nil, taskState),
		  data.NewField("deployedby", nil, deployedBy),
		  data.NewField("created", nil, created),
		  data.NewField("queuetime", nil, queueTime),
		  data.NewField("starttime", nil, startTime),*/
		data.NewField("duration", nil, duration))

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
	// Called before creatinga a new instance to allow plugin authors
	// to cleanup.
}
