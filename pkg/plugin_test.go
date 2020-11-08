package main

import (
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"os"
	"testing"
)

func TestConnection(t *testing.T) {
	request := backend.QueryDataRequest{
		PluginContext: backend.PluginContext{
			DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
				JSONData:                []byte("{\"Server\": \"" + os.Getenv("SERVER") + "\", \"SpaceId\": \"" + os.Getenv("SPACE") + "\"}"),
				DecryptedSecureJSONData: map[string]string{"apiKey": os.Getenv("APIKEY")},
			},
		},
		Queries: []backend.DataQuery{backend.DataQuery{
			JSON: []byte("{}"),
		}},
	}
	datasource := SampleDatasource{}
	datasource.QueryData(nil, &request)
}
