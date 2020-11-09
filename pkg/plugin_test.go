package main

import (
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"os"
	"testing"
	"time"
)

func TestConnection(t *testing.T) {
	start, _ := time.Parse("2006-01-02", "2020-01-01")
	end, _ := time.Parse("2006-01-02", "2021-01-01")

	request := backend.QueryDataRequest{
		PluginContext: backend.PluginContext{
			DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
				JSONData:                []byte("{\"Server\": \"" + os.Getenv("SERVER") + "\", \"SpaceId\": \"" + os.Getenv("SPACE") + "\", \"BucketDuration\": 60}"),
				DecryptedSecureJSONData: map[string]string{"apiKey": os.Getenv("APIKEY")},
			},
		},
		Queries: []backend.DataQuery{backend.DataQuery{
			JSON: []byte("{}"),
			TimeRange: struct {
				From time.Time
				To   time.Time
			}{From: start, To: end},
		}},
	}
	datasource := SampleDatasource{}
	datasource.QueryData(nil, &request)
}
