package main

import (
	"context"
	"encoding/json"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"strings"
)

func (td *SampleDatasource) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	log.DefaultLogger.Info("CallResource called")

	server, apiKey := getConnectionDetails(req.PluginContext)
	pathElements := strings.Split(req.Path, "/")

	if len(pathElements) < 2 {
		return nil
	}

	resourceType := pathElements[len(pathElements)-1]

	entities := map[string]string{}
	if resourceType == "spaces" {
		entities, _ = getAllResources(resourceType, server, "", apiKey)
	} else {
		spaceName := pathElements[len(pathElements)-2]
		spaces, _ := getSpaces(server, apiKey)
		entities, _ = getAllResources(resourceType, server, spaces[spaceName], apiKey)
	}

	entityNames := []string{}

	for k, _ := range entities {
		entityNames = append(entityNames, k)
	}

	response := backend.CallResourceResponse{}

	response.Status = 200
	response.Body, _ = json.Marshal(entityNames)

	sender.Send(&response)

	return nil
}
