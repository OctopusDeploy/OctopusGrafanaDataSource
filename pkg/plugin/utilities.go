package plugin

import (
	"errors"
	"strings"
	"time"
)

const releaseHistoryDateFormat = "2006-01-02T15:04:05"
const dateFormat = "2006-01-02T15:04:05.000-07:00"

func minInt64(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

func maxDuration(x, y time.Duration) time.Duration {
	if x > y {
		return x
	}
	return y
}

func parseTime(timeString string) time.Time {
	parsedTime, err := time.Parse(releaseHistoryDateFormat, timeString)
	if err == nil {
		return parsedTime
	}
	return time.Time{}
}

func arrayAverage(items []uint32) float32 {
	if len(items) == 0 {
		return 0
	}

	total := uint32(0)
	for i := 0; i < len(items); i++ {
		total += items[i]
	}
	return float32(total) / float32(len(items))
}

func arrayAverageDurationIgnoreZero(items []uint32) uint32 {
	total := uint32(0)
	count := uint32(0)
	for i := 0; i < len(items); i++ {
		if items[i] != 0 {
			total += items[i]
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return total / count
}

func empty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

func dateDiff(date1 string, date2 string) (time.Duration, error) {
	date1Parsed, err1 := time.Parse(releaseHistoryDateFormat, date1)
	date2Parsed, err2 := time.Parse(releaseHistoryDateFormat, date2)

	if err1 == nil && err2 == nil {
		return date1Parsed.Sub(date2Parsed), nil
	}

	return time.Duration(0), errors.New("failed to parse one or both dates")
}

func boolToInt(input bool) uint32 {
	bitSetVar := uint32(0)
	if input {
		bitSetVar = 1
	}
	return bitSetVar
}

func parseTimes(deployments *Deployments) {
	for i := 0; i < len(deployments.Deployments); i++ {
		parsedTime, err := time.Parse(releaseHistoryDateFormat, deployments.Deployments[i].StartTime)
		if err == nil {
			deployments.Deployments[i].StartTimeParsed = parsedTime
		}

		parsedTime, err = time.Parse(releaseHistoryDateFormat, deployments.Deployments[i].CompletedTime)
		if err == nil {
			deployments.Deployments[i].CompletedTimeParsed = parsedTime
		}
	}
}

// getTimeToSuccess will match failed deployments, find the next successful
// deployment and return the time between the two deployments in minutes. It
// returns 0 for successful deployments, or failed deployments that have not
// been followed by a successful deployment.
func getTimeToSuccess(deployment Deployment, deployments []Deployment, index int) uint32 {
	if deployment.TaskState == "Failed" {
		for index2 := index + 1; index2 < len(deployments); index2++ {
			d2 := deployments[index2]
			if d2.TaskState == "Success" &&
				d2.ChannelId == deployment.ChannelId &&
				d2.EnvironmentId == deployment.EnvironmentId &&
				d2.ProjectId == deployment.ProjectId &&
				d2.TenantId == deployment.TenantId {
				timeToRecovery, err := dateDiff(
					d2.CompletedTime,
					deployment.CompletedTime)
				if err == nil {
					return uint32(timeToRecovery / time.Minute)
				}
			}
		}
	}

	return 0
}

// includeDeployment will determine if a deployment record satisfies the
// current query filters.
func includeDeployment(qm *queryModel, deployment *Deployment) bool {
	// The reporting feed is shared across all queries in a request and spans
	// the widest of their time ranges, so every deployment must be checked
	// against this query's own range.
	if deployment.CompletedTimeParsed.Before(qm.Query.TimeRange.From) || deployment.CompletedTimeParsed.After(qm.Query.TimeRange.To) {
		return false
	}

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

	if !empty(qm.TaskState) && deployment.TaskState != qm.TaskState {
		return false
	}

	return true
}
