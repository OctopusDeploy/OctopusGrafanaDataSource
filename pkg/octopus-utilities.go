package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
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

func createRequest(url string, apiKey string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Octopus-ApiKey", apiKey)

	client := &http.Client{Timeout: time.Second * 100}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("Response code was " + strconv.Itoa(resp.StatusCode))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func resourceNameToId(resourceType string, path string, space string, apiKey string, resourceName string) (string, error) {
	url := path + "/api/" + space + "/" + resourceType + "/" + slugify(resourceName) + "?apikey=" + apiKey

	body, err := createRequest(url, apiKey)
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

// getAllResources calls the "all" API endpoint to return all available resources in a name to id map
func getAllResources(resourceType string, server string, space string, apiKey string) (map[string]string, error) {
	var url string

	if !empty(space) {
		url = server + "/api/" + space + "/" + resourceType + "/all"
	} else {
		url = server + "/api/" + resourceType + "/all"
	}

	body, err := createRequest(url, apiKey)
	if err != nil {
		return nil, err
	}

	var parsedResults []IdResource
	err = json.Unmarshal(body, &parsedResults)

	if err == nil {
		results := make(map[string]string)
		for _, r := range parsedResults {
			results[r.Name] = r.Id
		}
		return results, nil
	}

	return nil, err
}

// getRelease returns the details of a specific release
func getRelease(releaseId string, server string, space string, apiKey string) (Release, error) {
	var url string

	if !empty(space) {
		url = server + "/api/" + space + "/releases/" + releaseId
	} else {
		url = server + "/api/releases/" + releaseId
	}

	body, err := createRequest(url, apiKey)
	if err != nil {
		return Release{}, err
	}

	var parsedResults Release
	err = json.Unmarshal(body, &parsedResults)

	if err == nil {
		time, err := time.Parse(dateFormat, parsedResults.Assembled)
		if err == nil {
			parsedResults.AssembledDate = time
		}
		return parsedResults, nil
	}

	return Release{}, err
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
