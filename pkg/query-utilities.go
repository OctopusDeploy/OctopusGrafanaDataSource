package main

import (
	"encoding/json"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"sort"
	"time"
)

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

func getQueryModel(jsonData []byte) (queryModel, error) {
	// Unmarshal the json into our queryModel
	var qm queryModel

	err := json.Unmarshal(jsonData, &qm)
	return qm, err
}

// includeDeployment will determine if a deployment record satisfies the current filters
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

	if !empty(qm.TaskState) && deployment.TaskState != qm.TaskState {
		return false
	}

	return true
}
