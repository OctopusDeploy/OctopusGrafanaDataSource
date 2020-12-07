package main

import (
	"encoding/json"
	"encoding/xml"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/backend/resource/httpadapter"
	"net/http"
	"strings"
	"time"
)

// handleProjectsMapping returns a map of project names to ids as part of a resource call
func (ds *SampleDatasource) handleSpaceEntityMapping(rw http.ResponseWriter, req *http.Request, entityType string) {
	pluginContext := httpadapter.PluginConfigFromContext(req.Context())
	server, apiKey := getConnectionDetails(pluginContext)
	pathElements := strings.Split(req.URL.Path, "/")
	spaceId := ""
	if len(pathElements) == 2 {
		spaceId = pathElements[len(pathElements)-1]
	}
	entities, _ := getAllResources("spaces", server, spaceId, apiKey)
	json, _ := json.Marshal(entities)
	rw.Write(json)
}

// handleSpaces returns a list of all the space names as part of a resource call
func (td *SampleDatasource) handleSpaces(rw http.ResponseWriter, req *http.Request) {
	pluginContext := httpadapter.PluginConfigFromContext(req.Context())
	server, apiKey := getConnectionDetails(pluginContext)
	entities, _ := getSpaceResources(server, apiKey)
	json, _ := json.Marshal(entities)
	rw.Write(json)
}

// handleResources returns a list of entities names as part of a resource call
func (td *SampleDatasource) handleResources(rw http.ResponseWriter, req *http.Request) {
	pluginContext := httpadapter.PluginConfigFromContext(req.Context())
	server, apiKey := getConnectionDetails(pluginContext)

	pathElements := strings.Split(req.URL.Path, "/")

	entities := map[string]string{}
	resourceType := pathElements[len(pathElements)-1]
	space := pathElements[len(pathElements)-3]
	entities, _ = getAllResources(resourceType, server, space, apiKey)

	json, _ := json.Marshal(entities)
	rw.Write(json)
}

// handleResources returns a list of entities names as part of a resource call
func (td *SampleDatasource) handleDeploymentResources(rw http.ResponseWriter, req *http.Request) {
	pluginContext := httpadapter.PluginConfigFromContext(req.Context())
	server, apiKey := getConnectionDetails(pluginContext)
	projectId := req.URL.Query().Get("projectId")
	environmentId := req.URL.Query().Get("environmentId")

	pathElements := strings.Split(req.URL.Path, "/")

	var entities []PlainDeployment
	space := pathElements[len(pathElements)-2]
	entities, _ = getDeployments(server, space, apiKey, projectId, environmentId)

	json, _ := json.Marshal(entities)
	rw.Write(json)
}

// handleReportingRequest returns a list reporting deployments. It takes a request from the grafana frontend, calls
// the Octopus XML endpoint, processes the XML, and returns the results as JSON.
func (td *SampleDatasource) handleReportingRequest(rw http.ResponseWriter, req *http.Request) {
	pluginContext := httpadapter.PluginConfigFromContext(req.Context())
	server, apiKey := getConnectionDetails(pluginContext)

	pathElements := strings.Split(req.URL.Path, "/")
	spaceId := pathElements[len(pathElements)-3]
	projectId := req.URL.Query().Get("projectId")
	environmentId := req.URL.Query().Get("environmentId")
	earliestDate, _ := time.Parse(octopusDateFormat, req.URL.Query().Get("fromCompletedTime"))
	latestDate, _ := time.Parse(octopusDateFormat, req.URL.Query().Get("toCompletedTime"))

	query := buildReportingQueryUrl(server, spaceId, environmentId, projectId, earliestDate, latestDate)

	log.DefaultLogger.Info("Annotation project ID: " + req.URL.Query().Get("projectId"))
	log.DefaultLogger.Info("Annotation environment ID: " + req.URL.Query().Get("environmentId"))
	log.DefaultLogger.Info("Annotation from time: " + req.URL.Query().Get("fromCompletedTime"))
	log.DefaultLogger.Info("Annotation to time: " + req.URL.Query().Get("toCompletedTime"))

	// populate the data map with the results of the API query
	deployments := &Deployments{}
	xmlData, err := createRequest(query, apiKey)
	if err == nil {
		xml.Unmarshal(xmlData, deployments)
	}

	parseTimes(*deployments)

	// Return JSON to the front end
	json, _ := json.Marshal(deployments)
	rw.Write(json)
}
