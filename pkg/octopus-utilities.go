package main

import (
	"encoding/json"
	"errors"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func createRequest(url string, apiKey string) ([]byte, error) {
	log.DefaultLogger.Info("GET request to " + url)

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
		log.DefaultLogger.Error("GET request to " + url + " failed: " + err.Error())
		return nil, err
	}

	log.DefaultLogger.Debug("GET request to " + url + " responded with:")
	log.DefaultLogger.Debug(string(body[:]))

	return body, nil
}

func getResourceUrl(resourceType string, server string, space string) string {
	if !empty(space) && resourceType != "spaces" {
		if resourceType == "deployments" || resourceType == "releases" {
			return server + "/api/" + space + "/" + resourceType
		} else {
			return server + "/api/" + space + "/" + resourceType + "/all"
		}
	} else if resourceType == "deployments" || resourceType == "releases" {
		// deployments and releases are odd endpoints in that the default one returns all records, and there is no "/all" endpoint
		return server + "/api/" + resourceType
	} else {
		return server + "/api/" + resourceType + "/all"
	}
}

// getSpaceResources calls the "all" API endpoint to return all available resources in a name to id map
func getSpaceResources(server string, apiKey string) (map[string]string, error) {
	url := getResourceUrl("spaces", server, "")

	body, err := createRequest(url, apiKey)
	if err != nil {
		return nil, err
	}

	var parsedResults []SpaceResource
	err = json.Unmarshal(body, &parsedResults)

	if err == nil {
		results := make(map[string]string)
		for _, r := range parsedResults {
			results[r.Name] = r.Id
			// the default space is the unnamed space, identified as a single space
			if r.IsDefault {
				results[" "] = r.Id
			}
		}
		return results, nil
	}

	return nil, err
}

// getAllResources calls the "all" API endpoint to return all available resources in a name to id map
func getAllResources(resourceType string, server string, space string, apiKey string) (map[string]string, error) {
	url := getResourceUrl(resourceType, server, space)

	body, err := createRequest(url, apiKey)
	if err != nil {
		return nil, err
	}

	var parsedResults []BaseResource
	err = json.Unmarshal(body, &parsedResults)

	if err == nil {
		results := make(map[string]string)
		for _, r := range parsedResults {
			if !empty(r.Version) {
				results[r.Version] = r.Id
			} else {
				results[r.Name] = r.Id
			}
		}
		return results, nil
	}

	return nil, err
}

// getDeployments returns the a list of deployments
func getDeployments(server string, space string, apiKey string, projectId string, environmentId string) ([]PlainDeployment, error) {
	var url string

	if !empty(space) {
		url = server + "/api/" + space + "/deployments"
	} else {
		url = server + "/api/deployments"
	}

	url += "?projects=" + projectId + "&environments=" + environmentId

	body, err := createRequest(url, apiKey)
	if err != nil {
		return []PlainDeployment{}, err
	}

	var parsedResults PlainDeploymentItems
	err = json.Unmarshal(body, &parsedResults)

	if err == nil {
		for index := 0; index < len(parsedResults.Items); index++ {
			time, err := time.Parse(dateFormat, parsedResults.Items[index].Created)
			if err == nil {
				parsedResults.Items[index].CreatedParsed = time
			} else {
				log.DefaultLogger.Error("Failed to parse date " + parsedResults.Items[index].Created)
			}
		}
		return parsedResults.Items, nil
	}

	return []PlainDeployment{}, err
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

func buildReportingQueryUrl(server string, spaceId string, environmentId string, projectId string, earliestDate time.Time, latestDate time.Time) string {
	// the reporting endpoint is unique in that it returns XML
	query := ""

	// Build the Octopus API URL
	if empty(spaceId) {
		query = server + "/api/reporting/deployments/xml?" +
			"fromCompletedTime=" + url.QueryEscape(earliestDate.Format(octopusDateFormat)) +
			"&toCompletedTime=" + url.QueryEscape(latestDate.Format(octopusDateFormat))
	} else {
		query = server + "/api/" + spaceId + "/reporting/deployments/xml?" +
			"fromCompletedTime=" + url.QueryEscape(earliestDate.Format(octopusDateFormat)) +
			"&toCompletedTime=" + url.QueryEscape(latestDate.Format(octopusDateFormat))
	}

	// Filter server side on the project
	if !empty(projectId) {
		query += "&projectId=" + url.QueryEscape(projectId)
	}

	// Filter server side on the environment
	if !empty(environmentId) {
		query += "&environmentId=" + url.QueryEscape(environmentId)
	}

	return query
}
