package main

import (
	"context"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

func (td *SampleDatasource) queryResources(resourceType string, space string, ctx context.Context, req *backend.QueryDataRequest) (backend.DataResponse, error) {
	server, apiKey, _ := getConnectionDetails(req.PluginContext)
	entities, err := getAllResources(resourceType, server, space, apiKey)
	if err != nil {
		return backend.DataResponse{}, err
	}

	entityNames := []string{}

	for k, _ := range entities {
		entityNames = append(entityNames, k)
	}

	// create data frame response
	frame := data.NewFrame("response")

	frame.Fields = append(frame.Fields,
		data.NewField(resourceType, nil, entityNames))

	response := backend.DataResponse{}

	response.Frames = append(response.Frames, frame)

	return response, nil
}
