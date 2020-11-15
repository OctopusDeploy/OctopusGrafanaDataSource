package main

import (
	"context"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"time"
)

func (td *SampleDatasource) queryTable(ctx context.Context, query backend.DataQuery, deployments Deployments) backend.DataResponse {
	// Unmarshal the json into our queryModel
	var qm queryModel

	response := backend.DataResponse{}

	// Unmarshal the json into our queryModel
	qm, err := getQueryModel(query.JSON)
	if err != nil {
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
	thisTimeToRecovery := []uint32{}

	for index, d := range deployments.Deployments {
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
			thisTimeToRecovery = append(thisTimeToRecovery, getTimeToSuccess(d, deployments.Deployments, index))
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
		data.NewField("duration", nil, duration),
		data.NewField("timeToRecovery", nil, thisTimeToRecovery))

	// add the frames to the response
	response.Frames = append(response.Frames, frame)

	return response
}
