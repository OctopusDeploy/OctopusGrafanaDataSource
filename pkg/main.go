package main

import (
	"os"

	"github.com/OctopusDeploy/OctopusGrafanaDataSource/pkg/plugin"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

func main() {
	// Manage handles the lifecycle of datasource instances: one instance per
	// configured datasource, disposed and recreated when its settings change.
	if err := datasource.Manage("octopus-deploy-xmlfeed", plugin.NewDatasource, datasource.ManageOpts{}); err != nil {
		log.DefaultLogger.Error(err.Error())
		os.Exit(1)
	}
}
