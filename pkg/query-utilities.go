package main

import (
	"encoding/json"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"time"
)

// getQueryDetails returns the dates that we need to query Octopus for
func getQueryDetails(req *backend.QueryDataRequest) (time.Time, time.Time) {
	earliestDate := time.Time{}
	latestDate := time.Time{}

	for i := 0; i < len(req.Queries); i++ {
		if earliestDate.Equal(time.Time{}) || req.Queries[i].TimeRange.From.Before(earliestDate) {
			earliestDate = req.Queries[i].TimeRange.From
		}

		if latestDate.Equal(time.Time{}) || req.Queries[i].TimeRange.To.After(latestDate) {
			latestDate = req.Queries[i].TimeRange.To
		}
	}

	return earliestDate, latestDate
}

func getQueryModel(jsonData []byte) (queryModel, error) {
	// Unmarshal the json into our queryModel
	var qm queryModel

	err := json.Unmarshal(jsonData, &qm)
	return qm, err
}

// includeDeployment will determine if a deployment record satisfies the current filters
func includeDeployment(qm *queryModel, deployment *Deployment) bool {
	log.DefaultLogger.Info("Environment name query: " + qm.EnvironmentName)

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
