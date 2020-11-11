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
	"net/http"
	"net/url"
	"sort"
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

func getQueryDetails(req *backend.QueryDataRequest, path string, space string, apiKey string) (time.Time, time.Time, string, string) {
	earliestDate := time.Time{}
	latestDate := time.Time{}
	projects := []string{}
	environments := []string{}

	for i := 0; i < len(req.Queries); i++ {
		if earliestDate.Equal(time.Time{}) || req.Queries[i].TimeRange.From.Before(earliestDate) {
			earliestDate = req.Queries[i].TimeRange.From
		}

		if latestDate.Equal(time.Time{}) || req.Queries[i].TimeRange.To.After(latestDate) {
			latestDate = req.Queries[i].TimeRange.To
		}

		var qm queryModel
		response := backend.DataResponse{}

		response.Error = json.Unmarshal(req.Queries[i].JSON, &qm)
		if response.Error == nil {
			projects = append(projects, qm.ProjectName)
			environments = append(environments, qm.EnvironmentName)
		}
	}

	sort.Strings(projects)
	project := ""
	if projects[0] == projects[len(projects)-1] {
		projectName, err := resourceNameToId("projects", path, space, apiKey, projects[0])
		if err == nil {
			project = projectName
		}
	}

	return earliestDate, latestDate, project, ""
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

func includeDeployment(qm *queryModel, deployment *Deployment) bool {
	if !empty(qm.ReleaseVersion) && deployment.ReleaseVersion != qm.ReleaseVersion {
		return false
	}

	if !empty(qm.ProjectName) && deployment.ProjectName != qm.ProjectName {
		return false
	}

	if !empty(qm.ChannelName) && deployment.ChannelName != qm.ChannelName {
		return false
	}

	if !empty(qm.TenantName) && deployment.TenantName != qm.TenantName {
		return false
	}

	if !empty(qm.EnvironmentName) && deployment.EnvironmentName != qm.EnvironmentName {
		return false
	}

	return true
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
	avgDuration := []float32{}
	totalDuration := []uint32{}
	success := []uint32{}
	failure := []uint32{}
	cancelled := []uint32{}
	timedOut := []uint32{}
	totalTimeToRecovery := []uint32{}
	avgTimeToRecovery := []uint32{}

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
		bucketTotalTime := []uint32{}
		bucketTimeToRecovery := []uint32{}

		roundedTime := query.TimeRange.From.Add(bucketDuration * time.Duration(i)).Round(bucketDuration)
		if query.TimeRange.From.Before(roundedTime) && query.TimeRange.To.After(roundedTime) {

			count := 0

			// This could be optimised with some sorting and culling
			for index, d := range deployments.Deployments {
				if includeDeployment(&qm, &d) && d.CompetedTimeRounded.Equal(roundedTime) {

					count++

					// If this task was a failure, scan forward to the next success
					thisTimeToRecovery := uint32(0)
					if d.TaskState == "Failed" {
						for index2 := index + 1; index2 < len(deployments.Deployments); index2++ {
							d2 := deployments.Deployments[index2]
							if d2.TaskState == "Success" &&
								d2.ChannelId == d.ChannelId &&
								d2.EnvironmentId == d.EnvironmentId &&
								d2.ProjectId == d.ProjectId &&
								d2.TenantId == d.TenantId {
								timeToRecovery2, err := dateDiff(
									d2.CompletedTime,
									d.CompletedTime)
								if err == nil {
									thisTimeToRecovery = uint32(timeToRecovery2 / time.Minute)
								}
							}
						}
					}

					bucketTimeToRecovery = append(bucketTimeToRecovery, thisTimeToRecovery)
					bucketTotalTime = append(bucketTotalTime, d.DurationSeconds)

					if len(times) != 0 && times[len(times)-1].Equal(roundedTime) {
						success[len(success)-1] += func() uint32 {
							if d.TaskState == "Success" {
								return 1
							} else {
								return 0
							}
						}()
						failure[len(failure)-1] += func() uint32 {
							if d.TaskState == "Failed" {
								return 1
							} else {
								return 0
							}
						}()
						cancelled[len(cancelled)-1] += func() uint32 {
							if d.TaskState == "Cancelled" {
								return 1
							} else {
								return 0
							}
						}()
						timedOut[len(timedOut)-1] += func() uint32 {
							if d.TaskState == "TimedOut" {
								return 1
							} else {
								return 0
							}
						}()
						totalDuration[len(totalDuration)-1] += d.DurationSeconds
						avgDuration[len(avgDuration)-1] = arrayAverage(bucketTotalTime)
						totalTimeToRecovery[len(totalTimeToRecovery)-1] += thisTimeToRecovery
						avgTimeToRecovery[len(avgTimeToRecovery)-1] = arrayAverageDurationIgnoreZero(bucketTimeToRecovery)
					} else {
						times = append(times, roundedTime)
						success = append(success, func() uint32 {
							if d.TaskState == "Success" {
								return 1
							} else {
								return 0
							}
						}())
						failure = append(failure, func() uint32 {
							if d.TaskState == "Failed" {
								return 1
							} else {
								return 0
							}
						}())
						cancelled = append(cancelled, func() uint32 {
							if d.TaskState == "Cancelled" {
								return 1
							} else {
								return 0
							}
						}())
						timedOut = append(timedOut, func() uint32 {
							if d.TaskState == "TimedOut" {
								return 1
							} else {
								return 0
							}
						}())
						avgDuration = append(avgDuration, float32(d.DurationSeconds))
						totalDuration = append(totalDuration, d.DurationSeconds)
						totalTimeToRecovery = append(totalTimeToRecovery, thisTimeToRecovery)
						avgTimeToRecovery = append(avgTimeToRecovery, thisTimeToRecovery)
					}
				}
			}

			if count == 0 {
				times = append(times, roundedTime)
				success = append(success, 0)
				failure = append(failure, 0)
				cancelled = append(cancelled, 0)
				timedOut = append(timedOut, 0)
				avgDuration = append(avgDuration, 0)
				totalDuration = append(totalDuration, 0)
				totalTimeToRecovery = append(totalTimeToRecovery, 0)
				avgTimeToRecovery = append(avgTimeToRecovery, 0)
			}
		}
	}

	frame.Fields = append(frame.Fields, data.NewField("time", nil, times))

	if qm.SuccessField {
		frame.Fields = append(frame.Fields, data.NewField("success", nil, success))
	}

	if qm.FailureField {
		frame.Fields = append(frame.Fields, data.NewField("failure", nil, failure))
	}

	if qm.CancelledField {
		frame.Fields = append(frame.Fields, data.NewField("cancelled", nil, cancelled))
	}

	if qm.TimedOutField {
		frame.Fields = append(frame.Fields, data.NewField("timedOut", nil, timedOut))
	}

	if qm.TotalDurationField {
		frame.Fields = append(frame.Fields, data.NewField("totalDuration", nil, totalDuration))
	}

	if qm.AverageDurationField {
		frame.Fields = append(frame.Fields, data.NewField("avgDuration", nil, avgDuration))
	}

	if qm.TotalTimeToRecoveryField {
		frame.Fields = append(frame.Fields, data.NewField("totalTimToRecovery", nil, totalTimeToRecovery))
	}

	if qm.AverageTimeToRecoveryField {
		frame.Fields = append(frame.Fields, data.NewField("avgTimeToRecovery", nil, avgTimeToRecovery))
	}

	// add the frames to the response
	response.Frames = append(response.Frames, frame)

	return response
}

func (td *SampleDatasource) queryTable(ctx context.Context, query backend.DataQuery, deployments Deployments, bucketDuration time.Duration) backend.DataResponse {
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
	created := []time.Time{}
	queueTime := []time.Time{}
	startTime := []time.Time{}
	duration := []uint32{}

	for _, d := range deployments.Deployments {
		if includeDeployment(&qm, &d) {
			times = append(times, parseTime(d.CompletedTime))
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
			created = append(created, parseTime(d.Created))
			queueTime = append(queueTime, parseTime(d.QueueTime))
			startTime = append(startTime, parseTime(d.StartTime))
			duration = append(duration, d.DurationSeconds)
		}
	}

	frame.Fields = append(frame.Fields,
		data.NewField("time", nil, times),
		data.NewField("deploymentid", nil, deploymentId),
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
		data.NewField("starttime", nil, startTime),
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
	// Called before creating a new instance to allow plugin authors
	// to cleanup.
}
