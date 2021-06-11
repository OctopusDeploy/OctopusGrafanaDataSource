package main

import (
	"encoding/json"
	"encoding/xml"
	"github.com/grafana/grafana-plugin-sdk-go/backend/resource/httpadapter"
	"net/http"
	"strings"
	"time"
)

// Keeps the last range of deployments that were requested. This allows us to only query
// the new deployments that fall within the requested range, because calling
// /api/reporting/deployments/xml can be expensive.
var deploymentsCache = map[string][]Deployment{}

// handleProjectsMapping returns a map of project names to ids as part of a resource call
func (ds *SampleDatasource) handleSpaceEntityMapping(rw http.ResponseWriter, req *http.Request, entityType string) {
	pluginContext := httpadapter.PluginConfigFromContext(req.Context())
	server, apiKey, cacheDuration := getConnectionDetails(pluginContext)
	pathElements := strings.Split(req.URL.Path, "/")
	spaceId := ""
	if len(pathElements) == 2 {
		spaceId = pathElements[len(pathElements)-1]
	}
	entities, _ := getAllResources("spaces", server, spaceId, apiKey, cacheDuration)
	json, _ := json.Marshal(entities)
	rw.Write(json)
}

// handleSpaces returns a list of all the space names as part of a resource call
func (td *SampleDatasource) handleSpaces(rw http.ResponseWriter, req *http.Request) {
	pluginContext := httpadapter.PluginConfigFromContext(req.Context())
	server, apiKey, cacheDuration := getConnectionDetails(pluginContext)
	entities, _ := getSpaceResources(server, apiKey, cacheDuration)
	json, _ := json.Marshal(entities)
	rw.Write(json)
}

// handleResources returns a list of entities names as part of a resource call
func (td *SampleDatasource) handleResources(rw http.ResponseWriter, req *http.Request) {
	pluginContext := httpadapter.PluginConfigFromContext(req.Context())
	server, apiKey, cacheDuration := getConnectionDetails(pluginContext)

	pathElements := strings.Split(req.URL.Path, "/")

	entities := map[string]string{}
	resourceType := pathElements[len(pathElements)-1]
	space := pathElements[len(pathElements)-3]
	entities, _ = getAllResources(resourceType, server, space, apiKey, cacheDuration)

	json, _ := json.Marshal(entities)
	rw.Write(json)
}

// handleResources returns a list of entities names as part of a resource call
func (td *SampleDatasource) handleDeploymentResources(rw http.ResponseWriter, req *http.Request) {
	pluginContext := httpadapter.PluginConfigFromContext(req.Context())
	server, apiKey, cacheDuration := getConnectionDetails(pluginContext)
	projectId := req.URL.Query().Get("projectId")
	environmentId := req.URL.Query().Get("environmentId")

	pathElements := strings.Split(req.URL.Path, "/")

	var entities []PlainDeployment
	space := pathElements[len(pathElements)-2]
	entities, _ = getDeployments(server, space, apiKey, cacheDuration, projectId, environmentId)

	json, _ := json.Marshal(entities)
	rw.Write(json)
}

// handleReportingRequest returns a list reporting deployments. It takes a request from the grafana frontend, calls
// the Octopus XML endpoint, processes the XML, and returns the results as JSON.
func (td *SampleDatasource) handleReportingRequest(rw http.ResponseWriter, req *http.Request) {
	pluginContext := httpadapter.PluginConfigFromContext(req.Context())
	server, apiKey, cacheDuration := getConnectionDetails(pluginContext)

	pathElements := strings.Split(req.URL.Path, "/")
	spaceId := pathElements[len(pathElements)-3]
	projectId := req.URL.Query().Get("projectId")
	environmentId := req.URL.Query().Get("environmentId")
	earliestDate, _ := time.Parse(octopusDateFormat, req.URL.Query().Get("fromCompletedTime"))
	latestDate, _ := time.Parse(octopusDateFormat, req.URL.Query().Get("toCompletedTime"))

	if _, ok := deploymentsCache[spaceId]; !ok {
		deploymentsCache[spaceId] = []Deployment{}
	}

	// Prepend any deployments before the earliest cached record
	if len(deploymentsCache) != 0 && deploymentsCache[spaceId][0].StartTimeParsed.After(earliestDate) {
		query := buildReportingQueryUrl(server, spaceId, environmentId, projectId, earliestDate, deploymentsCache[spaceId][0].StartTimeParsed)
		deployments := getReturnAndProcessDeployments(query, apiKey, cacheDuration)
		deploymentsCache[spaceId] = append(deployments, deploymentsCache[spaceId]...)
	}

	// Append any deployments after the latest record
	if len(deploymentsCache) != 0 && deploymentsCache[spaceId][len(deploymentsCache)-1].CompletedTimeParsed.Before(latestDate) {
		query := buildReportingQueryUrl(server, spaceId, environmentId, projectId, deploymentsCache[spaceId][len(deploymentsCache)-1].CompletedTimeParsed, latestDate)
		deployments := getReturnAndProcessDeployments(query, apiKey, cacheDuration)
		deploymentsCache[spaceId] = append(deploymentsCache[spaceId], deployments...)
	}

	// Trim the cache to the new range
	deploymentsCache[spaceId] = returnDeploymentsWithinRange(deploymentsCache[spaceId], earliestDate, latestDate)

	deployment := Deployments{Deployments: deploymentsCache[spaceId]}

	// Return JSON to the front end
	json, _ := json.Marshal(deployment)
	rw.Write(json)
}

func getReturnAndProcessDeployments(query string, apiKey string, cacheDuration string) []Deployment {
	// populate the data map with the results of the API query
	deployments := &Deployments{}
	xmlData, err := createRequest(query, apiKey, cacheDuration)
	if err == nil {
		xml.Unmarshal(xmlData, deployments)
	}

	parseTimes(*deployments)

	return deployments.Deployments
}

func returnDeploymentsWithinRange(deployments []Deployment, startTime time.Time, endTime time.Time) []Deployment {
	values := []Deployment{}
	for i := range deployments {
		if !startTime.After(deployments[i].StartTimeParsed) && !endTime.Before(deployments[i].CompletedTimeParsed) {
			// A sanity check to weed out duplicate deployment IDs. This catches the possibility
			// that some query dates exactly overlapped.
			if !deploymentsContains(values, deployments[i].DeploymentId) {
				values = append(values, deployments[i])
			}
		}
	}
	return values
}

func deploymentsContains(deployments []Deployment, deploymentId string) bool {
	for _, d := range deployments {
		if d.DeploymentId == deploymentId {
			return true
		}
	}
	return false
}
