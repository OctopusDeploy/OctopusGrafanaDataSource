# Octopus Deploy Grafana Datasource

[This repo](https://github.com/OctopusDeploy/OctopusGrafanaDataSource) holds the source code to the Octopus Deploy Grafana datasource plugin.

The plugin connects to the [reporting](https://octopus.com/docs/administration/reporting) endpoint at http://octopuserver/api/reporting/deployments/xml and converts the results to a time series that can be displayed in graphs, or as a table to be displayed in a grafana table.

Other entities such as environments, projects, tenants etc. are also exposed as tables.

This plugin requires Grafana 7 or later.

![image](https://user-images.githubusercontent.com/160104/99312386-b10dfc80-28a9-11eb-98e7-3324c222b392.png)

# Support

This plugin is released as an early access. We expect it has bugs and gaps in functionality, and is only recommended for testing.

If you have found an issue or have a suggestion, please reach out to our [support teams](https://octopus.com/support).

# Download

The plugin can be downloaded from the [GitHub Releases](https://github.com/OctopusDeploy/OctopusGrafanaDataSource/releases) page.

This ZIP file is then extracted into the Grafana plugin directory (which is usually `INSTALL_DIR\data\plugins` or `/var/lib/grafana/plugins`):

```
unzip octopus_grafana_datasource.zip -d YOUR_PLUGIN_DIR/octopus
```

See the [Grafana documentation](https://grafana.com/docs/grafana/latest/plugins/installation/#install-a-packaged-plugin) for more details.

# Signing

This plugin is unsigned, meaning the plugin must be listed in the `GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS` environment variable (e.g `GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=octopus-deploy-xmlfeed`) or the `allow_loading_unsigned_plugins` option in `grafana.ini` must list `octopus-deploy-xmlfeed` e.g.:

```ini
[plugins]
allow_loading_unsigned_plugins = octopus-deploy-xmlfeed
```

See the [Grafana documentation](https://grafana.com/docs/grafana/latest/administration/configuration/#allow_loading_unsigned_plugins) for more details.

# Octopus Permissions

The account used to query Octopus requires the following permissions in the spaces that Grafana will report on:

* DeploymentView
* EnvironmentView
* TenantView
* ProcessView
* ProjectView
* ReleaseView

# Building

The following tools are required to build the plugin:

* [Go](https://golang.org/dl/)
* [Mage](https://magefile.org/#installation)
* [Nodejs](https://nodejs.org/en/download/)
* [Yarn](https://classic.yarnpkg.com/en/docs/install)

Build the plugin with:

```
yarn build
mage -v
```

# Proxy support

The backend plugin respects the `HTTP_PROXY`, `HTTPS_PROXY`, and `NO_PROXY` environment variables. The [go documentation](https://pkg.go.dev/golang.org/x/net/http/httpproxy#FromEnvironment)
describes the format of these variables.

HTTPS proxies with custom certificates must embed the CA cert to work correctly. 

The `Dockerfile` below demonstrates how to add a custom certificate:

```yaml
FROM octopussamples/grafana

USER root
COPY octo.domain.local.crt /usr/local/share/ca-certificates/
RUN update-ca-certificates
```

# GitHub Actions

This project is built and published via [GitHub Actions](https://github.com/OctopusDeploy/OctopusGrafanaDataSource/actions).

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

![image](https://user-images.githubusercontent.com/160104/99312462-d13dbb80-28a9-11eb-9977-1fc89c3348b0.png)

# Caching

Calling the Octopus API endpoints like /api/reporting/deployments/xml can be expensive, especially if there are many deployments to return and the Grafana date range is quite large.

The plugin will cache results from /api/reporting/deployments/xml to improve performace. The first request will return all the results, but subsequent requests will only query Octopus for results before and after those that were cached. So a Grafana dashboard set to refresh every 5 minutes will result in queries to Octopus for the last 5 minutes worth of data.

The datasource also exposes a field to define a cache duration. This applies to entities like projects, environments, channels etc. The cache duration can be left blank, in which case all these entities are requested from Octopus every time. Setting a duration can improve performance where many people are viewing the same dashboard, as only the first request will require an API call to Octopus, and others will share the same result.

## Stats

![Github All Releases](https://img.shields.io/github/downloads/OctopusDeploy/OctopusGrafanaDataSource/total.svg)
