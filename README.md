# Octopus Deploy Grafana Datasource

[This repo](https://github.com/OctopusDeploy/OctopusGrafanaDataSource) holds the source code to the Octopus Deploy Grafana datasource plugin.

The plugin connects to the [reporting](https://octopus.com/docs/administration/reporting) endpoint at `https://octopusserver/api/reporting/deployments/xml` and converts the results to a time series that can be displayed in graphs, or as a table to be displayed in a Grafana table.

Other entities such as environments, projects, tenants etc. are also exposed as tables, and deployments can be overlaid on any panel as annotations.

This plugin requires Grafana 10.4 or later and is tested against Grafana 11.x and 12.x.

![image](https://user-images.githubusercontent.com/160104/99312386-b10dfc80-28a9-11eb-98e7-3324c222b392.png)

# Support

If you have found an issue or have a suggestion, please reach out to our [support teams](https://octopus.com/support).

# Download

The plugin can be downloaded from the [GitHub Releases](https://github.com/OctopusDeploy/OctopusGrafanaDataSource/releases) page.

The ZIP file is extracted into the Grafana plugin directory (usually `INSTALL_DIR\data\plugins` or `/var/lib/grafana/plugins`):

```
unzip octopus-deploy-xmlfeed-<version>.zip -d YOUR_PLUGIN_DIR
```

See the [Grafana documentation](https://grafana.com/docs/grafana/latest/administration/plugin-management/#install-plugin-on-local-grafana) for more details.

# Signing

This plugin is unsigned, meaning the plugin must be listed in the `GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS` environment variable (e.g `GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=octopus-deploy-xmlfeed`) or the `allow_loading_unsigned_plugins` option in `grafana.ini` must list `octopus-deploy-xmlfeed` e.g.:

```ini
[plugins]
allow_loading_unsigned_plugins = octopus-deploy-xmlfeed
```

See the [Grafana documentation](https://grafana.com/docs/grafana/latest/setup-grafana/configure-grafana/#allow_loading_unsigned_plugins) for more details.

If your organisation does not allow unsigned plugins, you can sign the release ZIP yourself for your own Grafana instances with a free Grafana Cloud access policy token: extract the ZIP and run `npx @grafana/sign-plugin --rootUrls https://your-grafana-instance/` in the extracted directory. See [Sign a plugin](https://grafana.com/developers/plugin-tools/publish-a-plugin/sign-a-plugin) for details on generating the token.

# Octopus Permissions

The account used to query Octopus requires the following permissions in the spaces that Grafana will report on:

* DeploymentView
* EnvironmentView
* TenantView
* ProcessView
* ProjectView
* ReleaseView

Consider creating a [service account](https://octopus.com/docs/security/users-and-teams/service-accounts) with only these permissions and generating the API key from it.

# Building

The following tools are required to build the plugin:

* [Go](https://go.dev/dl/) 1.26 or later
* [Mage](https://magefile.org/#installation)
* [Node.js](https://nodejs.org/en/download/) 22 or later

Build the plugin with:

```
npm install
npm run build
mage -v
```

Run the test suites with:

```
npm run test:ci
go test ./pkg/...
```

# Development environment

A Grafana instance with the plugin installed can be started with Docker:

```
npm run server
```

The provisioned datasource reads the Octopus server URL and API key from the `OCTOPUS_SERVER` and `OCTOPUS_API_KEY` environment variables.

# Proxy support

The backend plugin respects the `HTTP_PROXY`, `HTTPS_PROXY`, and `NO_PROXY` environment variables. The [go documentation](https://pkg.go.dev/golang.org/x/net/http/httpproxy#FromEnvironment)
describes the format of these variables.

HTTPS proxies with custom certificates must have the CA certificate installed in the operating system certificate store to work correctly.

# GitHub Actions

This project is built and tested via [GitHub Actions](https://github.com/OctopusDeploy/OctopusGrafanaDataSource/actions). Releases are created by pushing a `v*` tag.

# Sample Dashboard

A sample dashboard displaying data returned by this plugin can be found on the [Grafana Dashboard Gallery](https://grafana.com/grafana/dashboards/13413).

![image](https://user-images.githubusercontent.com/160104/99312462-d13dbb80-28a9-11eb-9977-1fc89c3348b0.png)

# DORA metrics

Octopus Deploy 2022.3 and later calculates DORA metrics natively via the [Insights](https://octopus.com/docs/insights) feature. The lead time and time to recovery fields exposed by this plugin are approximations derived from the reporting feed; prefer Insights where it is available.

# Caching

Calling the Octopus API endpoints like `/api/reporting/deployments/xml` can be expensive, especially if there are many deployments to return and the Grafana date range is quite large. The plugin caches reporting feed responses for their exact time window, and release details for 24 hours, as this data does not change once written.

The datasource also exposes a field to define a cache duration. This applies to entities like projects, environments, channels etc. The cache duration can be left blank, in which case these entities are requested from Octopus every time. Setting a duration can improve performance where many people are viewing the same dashboard.

All caches are scoped to an individual datasource instance, so data is never shared between datasources with different servers or credentials.
