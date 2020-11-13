package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
)

func slugify(value string) string {
	value = strings.ToLower(value)
	value = regexp.MustCompile(`\s`).ReplaceAllString(value, "-")
	value = regexp.MustCompile(`[^a-zA-Z0-9-]`).ReplaceAllString(value, "-")
	value = regexp.MustCompile(`-+`).ReplaceAllString(value, "-")
	value = strings.Trim(value, "-/")
	return value
}

func resourceNameToId(resourceType string, path string, space string, apiKey string, resourceName string) (string, error) {
	url := path + "/api/" + space + "/" + resourceType + "/" + slugify(resourceName) + "?apikey=" + apiKey
	resp, err := http.Get(url)
	defer resp.Body.Close()

	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	parsedResults := IdResource{}
	err = json.Unmarshal(body, &parsedResults)

	if err == nil {
		return parsedResults.Id, nil
	}

	return "", err
}

// getTimeToSuccess will match failed deployments, find the next successful deployment
// and return the time between the two deployments. It returns 0 for successful deployments,
// or failed deployments that have not been followed by a successful deployment.
func getTimeToSuccess(deployment Deployment, deployments []Deployment, index int) uint32 {
	// If this task was a failure, scan forward to the next success
	if deployment.TaskState == "Failed" {
		for index2 := index + 1; index2 < len(deployments); index2++ {
			d2 := deployments[index2]
			if d2.TaskState == "Success" &&
				d2.ChannelId == deployment.ChannelId &&
				d2.EnvironmentId == deployment.EnvironmentId &&
				d2.ProjectId == deployment.ProjectId &&
				d2.TenantId == deployment.TenantId {
				timeToRecovery2, err := dateDiff(
					d2.CompletedTime,
					deployment.CompletedTime)
				if err == nil {
					return uint32(timeToRecovery2 / time.Minute)
				}
			}
		}
	}

	return 0
}
