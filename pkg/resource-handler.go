package main

import (
	"encoding/json"
	"github.com/grafana/grafana-plugin-sdk-go/backend/resource/httpadapter"
	"net/http"
	"strings"
)

func (ds *SampleDatasource) handleSpaces(rw http.ResponseWriter, req *http.Request) {
	pluginContext := httpadapter.PluginConfigFromContext(req.Context())
	server, apiKey := getConnectionDetails(pluginContext)
	entities, _ := getAllResources("spaces", server, "", apiKey)
	json, _ := json.Marshal(entities)
	rw.Write(json)
}

func (td *SampleDatasource) handleResources(rw http.ResponseWriter, req *http.Request) {
	pluginContext := httpadapter.PluginConfigFromContext(req.Context())
	server, apiKey := getConnectionDetails(pluginContext)

	pathElements := strings.Split(req.URL.Path, "/")

	entities := map[string]string{}
	resourceType := pathElements[len(pathElements)-1]
	space := pathElements[len(pathElements)-2]
	entities, _ = getAllResources(resourceType, server, space, apiKey)

	entityNames := []string{}

	for k, _ := range entities {
		entityNames = append(entityNames, k)
	}

	json, _ := json.Marshal(entityNames)
	rw.Write(json)
}
