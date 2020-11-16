package main

import (
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

func (td *SampleDatasource) queryResources(entities map[string]string, resourceType string) (backend.DataResponse, error) {
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
