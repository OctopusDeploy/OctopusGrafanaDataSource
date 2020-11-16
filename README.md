# Octopus Deploy Grafana Datasource

This repo holds the source code to the Octopus Deploy Grafana datasource plugin.

The plugin connects to the [reporting](https://octopus.com/docs/administration/reporting) endpoint at http://octopuserver/api/reporting/deployments/xml and converts the results to a time series that can be displayed in graphs, or as a table to be displayed in a grafana table.

Other entities such as environments, projects, tenants etc. are also exposed as tables.

# Building

The following tools are required to build the plugin:

* [Go](https://golang.org/dl/)
* [Mage](https://magefile.org/#installation)
* [Nodejs](https://nodejs.org/en/download/)
* [Yarn](https://classic.yarnpkg.com/en/docs/install/#windows-stable)

Build the plugin with:

```
yarn build
mage -v
```

# Docker

The plugin can be run with the Grafana Docker image with the command:

```
docker run -e "GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=octopus-deploy-xmlfeed" -d -p 3000:3000 -v "$(pwd):/var/lib/grafana/plugins" --name=grafana grafana/grafana:7.0.0
```

A docker image with the plugin already installed and the `GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS` setting configured can be found in [Docker Hub](https://hub.docker.com/r/octopussamples/grafana), and run with the command:

```
docker run -d -p 3000:3000 --name=grafana octopussamples/grafana:latest
```

# Sample Dashboard

A sample dashboard displaying data returned by this plugin can be found on the [Grafana Dashboard Gallery](https://grafana.com/grafana/dashboards/13413).
