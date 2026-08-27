package plugin

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// maxFrames caps the number of time buckets in a time series response.
const maxFrames = 50

func getBucketDuration(queryDuration time.Duration, bucketDuration time.Duration) (int64, time.Duration) {
	buckets := minInt64(maxFrames, int64(queryDuration/bucketDuration))
	if buckets < 1 {
		buckets = 1
	}
	return buckets, queryDuration / time.Duration(buckets)
}

func setCompletedTimeRounded(deployments Deployments, bucketDuration time.Duration) {
	for i := 0; i < len(deployments.Deployments); i++ {
		parsed, err := time.Parse(releaseHistoryDateFormat, deployments.Deployments[i].CompletedTime)
		if err == nil {
			deployments.Deployments[i].CompletedTimeRounded = parsed.Round(bucketDuration)
		}
	}
}

// queryTimeSeries generates a time series response, combining deployment
// information into time buckets that can be displayed in a graph.
func (d *Datasource) queryTimeSeries(ctx context.Context, qm queryModel, spaceID string, deployments Deployments) backend.DataResponse {
	response := backend.DataResponse{}
	query := qm.Query

	frame := data.NewFrame("response")

	times := []time.Time{}
	avgDuration := []float32{}
	totalDuration := []uint32{}
	success := []uint32{}
	failure := []uint32{}
	cancelled := []uint32{}
	timedOut := []uint32{}
	totalTimeToRecovery := []uint32{}
	avgTimeToRecovery := []uint32{}
	totalCycleTime := []uint32{}
	avgCycleTime := []uint32{}

	maxDataPoints := query.MaxDataPoints
	if maxDataPoints < 1 {
		maxDataPoints = maxFrames
	}

	// Work out how long the buckets should be
	buckets, bucketDuration := getBucketDuration(query.TimeRange.Duration(), time.Duration(int64(query.TimeRange.Duration())/maxDataPoints))

	// get the bucket start time for each deployment
	setCompletedTimeRounded(deployments, bucketDuration)

	for i := 0; i < int(buckets); i++ {
		bucketTotalTime := []uint32{}
		bucketTimeToRecovery := []uint32{}
		bucketCycleTime := []uint32{}

		// Get the time that starts this bucket
		roundedTime := query.TimeRange.From.Add(bucketDuration * time.Duration(i)).Round(bucketDuration)

		// Grafana really doesn't like it if you have records outside of the
		// range, so make sure we are definitely inside the query range here.
		if query.TimeRange.From.Before(roundedTime) && query.TimeRange.To.After(roundedTime) {

			count := 0

			for index, deployment := range deployments.Deployments {
				// Make sure the deployment matches the query filters, and the
				// deployment completion time matches the start of this bucket
				if includeDeployment(&qm, &deployment) && deployment.CompletedTimeRounded.Equal(roundedTime) {

					thisCycleTime := uint32(0)

					// Don't make the extra API calls if we don't need to
					if qm.AverageCycleTimeField || qm.TotalCycleTimeField {
						// Get the time from when the release was created. This
						// is only available while the release is still in the
						// database, as the release creation date is not stored
						// by the reporting endpoint.
						releaseDetails, err := d.client.getRelease(ctx, spaceID, deployment.ReleaseId)

						if err == nil {
							diff := parseTime(deployment.CompletedTime).Sub(releaseDetails.AssembledDate).Seconds()
							bucketCycleTime = append(bucketCycleTime, uint32(diff))
							thisCycleTime = uint32(diff)
						}
					}

					count++

					// If this task was a failure, scan forward to the next success
					thisTimeToRecovery := getTimeToSuccess(deployment, deployments.Deployments, index)

					bucketTimeToRecovery = append(bucketTimeToRecovery, thisTimeToRecovery)
					bucketTotalTime = append(bucketTotalTime, deployment.DurationSeconds)

					if len(times) != 0 && times[len(times)-1].Equal(roundedTime) {
						success[len(success)-1] += boolToInt(deployment.TaskState == "Success")
						failure[len(failure)-1] += boolToInt(deployment.TaskState == "Failed")
						cancelled[len(cancelled)-1] += boolToInt(deployment.TaskState == "Cancelled")
						timedOut[len(timedOut)-1] += boolToInt(deployment.TaskState == "TimedOut")
						totalDuration[len(totalDuration)-1] += deployment.DurationSeconds
						avgDuration[len(avgDuration)-1] = arrayAverage(bucketTotalTime)
						totalTimeToRecovery[len(totalTimeToRecovery)-1] += thisTimeToRecovery
						avgTimeToRecovery[len(avgTimeToRecovery)-1] = arrayAverageDurationIgnoreZero(bucketTimeToRecovery)
						avgCycleTime[len(avgCycleTime)-1] = arrayAverageDurationIgnoreZero(bucketCycleTime)
						totalCycleTime[len(totalCycleTime)-1] += thisCycleTime
					} else {
						times = append(times, roundedTime)
						success = append(success, boolToInt(deployment.TaskState == "Success"))
						failure = append(failure, boolToInt(deployment.TaskState == "Failed"))
						cancelled = append(cancelled, boolToInt(deployment.TaskState == "Cancelled"))
						timedOut = append(timedOut, boolToInt(deployment.TaskState == "TimedOut"))
						avgDuration = append(avgDuration, float32(deployment.DurationSeconds))
						totalDuration = append(totalDuration, deployment.DurationSeconds)
						totalTimeToRecovery = append(totalTimeToRecovery, thisTimeToRecovery)
						avgTimeToRecovery = append(avgTimeToRecovery, thisTimeToRecovery)
						avgCycleTime = append(avgCycleTime, thisCycleTime)
						totalCycleTime = append(totalCycleTime, thisCycleTime)
					}
				}
			}

			// If no deployments fell inside this time bucket, add a zero record
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
				avgCycleTime = append(avgCycleTime, 0)
				totalCycleTime = append(totalCycleTime, 0)
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
		frame.Fields = append(frame.Fields, data.NewField("totalTimeToRecovery", nil, totalTimeToRecovery))
	}

	if qm.AverageTimeToRecoveryField {
		frame.Fields = append(frame.Fields, data.NewField("avgTimeToRecovery", nil, avgTimeToRecovery))
	}

	if qm.TotalCycleTimeField {
		frame.Fields = append(frame.Fields, data.NewField("totalReleaseLeadTime", nil, totalCycleTime))
	}

	if qm.AverageCycleTimeField {
		frame.Fields = append(frame.Fields, data.NewField("avgReleaseLeadTime", nil, avgCycleTime))
	}

	response.Frames = append(response.Frames, frame)

	return response
}

// queryTable returns the deployments from the reporting feed as a table.
func (d *Datasource) queryTable(qm queryModel, deployments Deployments) backend.DataResponse {
	response := backend.DataResponse{}

	frame := data.NewFrame("response")

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
	timeToRecovery := []uint32{}

	for index, deployment := range deployments.Deployments {
		if includeDeployment(&qm, &deployment) {
			times = append(times, parseTime(deployment.CompletedTime))
			deploymentId = append(deploymentId, deployment.DeploymentId)
			deploymentName = append(deploymentName, deployment.DeploymentName)
			projectId = append(projectId, deployment.ProjectId)
			projectName = append(projectName, deployment.ProjectName)
			projectSlug = append(projectSlug, deployment.ProjectSlug)
			tenantId = append(tenantId, deployment.TenantId)
			tenantName = append(tenantName, deployment.TenantName)
			channelId = append(channelId, deployment.ChannelId)
			channelName = append(channelName, deployment.ChannelName)
			environmentId = append(environmentId, deployment.EnvironmentId)
			environmentName = append(environmentName, deployment.EnvironmentName)
			releaseId = append(releaseId, deployment.ReleaseId)
			releaseVersion = append(releaseVersion, deployment.ReleaseVersion)
			taskId = append(taskId, deployment.TaskId)
			taskState = append(taskState, deployment.TaskState)
			deployedBy = append(deployedBy, deployment.DeployedBy)
			created = append(created, parseTime(deployment.Created))
			queueTime = append(queueTime, parseTime(deployment.QueueTime))
			startTime = append(startTime, parseTime(deployment.StartTime))
			duration = append(duration, deployment.DurationSeconds)
			timeToRecovery = append(timeToRecovery, getTimeToSuccess(deployment, deployments.Deployments, index))
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
		data.NewField("timeToRecovery", nil, timeToRecovery))

	response.Frames = append(response.Frames, frame)

	return response
}

// queryResources returns the names of a resource collection as a single
// column table, sorted for a stable output.
func queryResources(entities map[string]string, resourceType string) backend.DataResponse {
	entityNames := []string{}

	for name := range entities {
		entityNames = append(entityNames, name)
	}
	sort.Strings(entityNames)

	frame := data.NewFrame("response")
	frame.Fields = append(frame.Fields,
		data.NewField(resourceType, nil, entityNames))

	response := backend.DataResponse{}
	response.Frames = append(response.Frames, frame)

	return response
}

// queryAnnotationReport converts reporting feed deployments into an
// annotation frame. Grafana maps the time, timeEnd, title, text and tags
// fields onto annotation events; tags are comma separated.
func (d *Datasource) queryAnnotationReport(qm queryModel, deployments Deployments) backend.DataResponse {
	times := []time.Time{}
	timeEnds := []time.Time{}
	titles := []string{}
	texts := []string{}
	tags := []string{}

	for _, deployment := range deployments.Deployments {
		if !includeDeployment(&qm, &deployment) {
			continue
		}

		times = append(times, deployment.StartTimeParsed)
		timeEnds = append(timeEnds, deployment.CompletedTimeParsed)
		titles = append(titles, strings.TrimSpace(deployment.ProjectName+" "+deployment.ReleaseVersion))
		texts = append(texts, deployment.DeploymentId)
		tags = append(tags, buildAnnotationTags(deployment))
	}

	frame := data.NewFrame("annotations",
		data.NewField("time", nil, times),
		data.NewField("timeEnd", nil, timeEnds),
		data.NewField("title", nil, titles),
		data.NewField("text", nil, texts),
		data.NewField("tags", nil, tags))

	response := backend.DataResponse{}
	response.Frames = append(response.Frames, frame)
	return response
}

func buildAnnotationTags(deployment Deployment) string {
	parts := []string{}
	if !empty(deployment.ProjectName) {
		parts = append(parts, "Project: "+deployment.ProjectName)
	}
	if !empty(deployment.TenantName) {
		parts = append(parts, "Tenant: "+deployment.TenantName)
	}
	if !empty(deployment.ChannelName) {
		parts = append(parts, "Channel: "+deployment.ChannelName)
	}
	if !empty(deployment.EnvironmentName) {
		parts = append(parts, "Environment: "+deployment.EnvironmentName)
	}
	if !empty(deployment.ReleaseVersion) {
		parts = append(parts, "Version: "+deployment.ReleaseVersion)
	}
	return strings.Join(parts, ",")
}

// queryAnnotationDeployments converts deployments from the JSON API into an
// annotation frame of point-in-time events.
func queryAnnotationDeployments(deployments []PlainDeployment) backend.DataResponse {
	times := []time.Time{}
	titles := []string{}
	texts := []string{}

	for _, deployment := range deployments {
		times = append(times, deployment.CreatedParsed)
		titles = append(titles, deployment.Name)
		texts = append(texts, deployment.Id)
	}

	frame := data.NewFrame("annotations",
		data.NewField("time", nil, times),
		data.NewField("title", nil, titles),
		data.NewField("text", nil, texts))

	response := backend.DataResponse{}
	response.Frames = append(response.Frames, frame)
	return response
}
