package main

import (
	"context"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"time"
)

func getBucketDuration(queryDuration time.Duration, bucketDuration time.Duration) (int64, time.Duration) {
	fixedDuration := time.Duration(60)
	if bucketDuration != 0 {
		fixedDuration = bucketDuration
	}
	buckets := Min(maxFrames, int64(queryDuration/fixedDuration))
	return buckets, queryDuration / time.Duration(buckets)
}

func setCompletedTimeRounded(deployments Deployments, bucketDuration time.Duration) {
	for i := 0; i < len(deployments.Deployments); i++ {
		time, err := time.Parse(releaseHistoryDateFormat, deployments.Deployments[i].CompletedTime)
		if err == nil {
			deployments.Deployments[i].CompetedTimeRounded = time.Round(bucketDuration)
		}
	}
}

// query generates a time series response, combining deployment information into time buckets
// that can be displayed in a graph.
func (td *SampleDatasource) query(ctx context.Context, query backend.DataQuery, deployments Deployments, server string, space string, apiKey string) backend.DataResponse {
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

	// Work out how long the buckets should be
	buckets, bucketDuration := getBucketDuration(query.TimeRange.Duration(), query.Interval)

	// get the bucket start time for each deployment
	setCompletedTimeRounded(deployments, bucketDuration)

	for i := 0; i < int(buckets); i++ {
		bucketTotalTime := []uint32{}
		bucketTimeToRecovery := []uint32{}
		bucketCycleTime := []uint32{}

		// Get the time that starts this bucket
		roundedTime := query.TimeRange.From.Add(bucketDuration * time.Duration(i)).Round(bucketDuration)

		// Grafana really doesn't like it if you have records outside of the range, so make
		// sure we are definitely inside the query range here.
		if query.TimeRange.From.Before(roundedTime) && query.TimeRange.To.After(roundedTime) {

			count := 0

			// This could be optimised with some sorting and culling
			for index, d := range deployments.Deployments {
				// Make sure the deployment matches the query filters, and the deployment
				// completion time matches the start of this time bucket
				if includeDeployment(&qm, &d) && d.CompetedTimeRounded.Equal(roundedTime) {

					thisCycleTime := uint32(0)

					// Don't make the extra API calls if we don't need to
					if qm.AverageCycleTimeField || qm.TotalCycleTimeField {
						// get the cycle time, or the time from when the release was created.
						// note we can only get this information if the release is still in the database, as the release creation
						// date is not stored by the reporting endpoint
						releaseDetails, err := getRelease(d.ReleaseId, server, space, apiKey)

						if err == nil {
							diff := parseTime(d.CompletedTime).Sub(releaseDetails.AssembledDate).Seconds()
							bucketCycleTime = append(bucketCycleTime, uint32(diff))
							thisCycleTime = uint32(diff)
						}
					}

					count++

					// If this task was a failure, scan forward to the next success
					thisTimeToRecovery := getTimeToSuccess(d, deployments.Deployments, index)

					bucketTimeToRecovery = append(bucketTimeToRecovery, thisTimeToRecovery)
					bucketTotalTime = append(bucketTotalTime, d.DurationSeconds)

					if len(times) != 0 && times[len(times)-1].Equal(roundedTime) {
						success[len(success)-1] += boolToInt(d.TaskState == "Success")
						failure[len(failure)-1] += boolToInt(d.TaskState == "Failed")
						cancelled[len(cancelled)-1] += boolToInt(d.TaskState == "Cancelled")
						timedOut[len(timedOut)-1] += boolToInt(d.TaskState == "TimedOut")
						totalDuration[len(totalDuration)-1] += d.DurationSeconds
						avgDuration[len(avgDuration)-1] = arrayAverage(bucketTotalTime)
						totalTimeToRecovery[len(totalTimeToRecovery)-1] += thisTimeToRecovery
						avgTimeToRecovery[len(avgTimeToRecovery)-1] = arrayAverageDurationIgnoreZero(bucketTimeToRecovery)
						avgCycleTime[len(avgCycleTime)-1] = arrayAverageDurationIgnoreZero(bucketCycleTime)
						totalCycleTime[len(totalCycleTime)-1] += thisCycleTime
					} else {
						times = append(times, roundedTime)
						success = append(success, boolToInt(d.TaskState == "Success"))
						failure = append(failure, boolToInt(d.TaskState == "Failed"))
						cancelled = append(cancelled, boolToInt(d.TaskState == "Cancelled"))
						timedOut = append(timedOut, boolToInt(d.TaskState == "TimedOut"))
						avgDuration = append(avgDuration, float32(d.DurationSeconds))
						totalDuration = append(totalDuration, d.DurationSeconds)
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
		frame.Fields = append(frame.Fields, data.NewField("totalCycleTime", nil, totalCycleTime))
	}

	if qm.AverageCycleTimeField {
		frame.Fields = append(frame.Fields, data.NewField("avgCycleTime", nil, avgCycleTime))
	}

	// add the frames to the response
	response.Frames = append(response.Frames, frame)

	return response
}
