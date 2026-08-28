# Changelog

## 2.0.0

Modernised the plugin for current Grafana releases.

- Requires Grafana 10.4 or later; tested against Grafana 11.x and 12.x.
- Replaced the AngularJS annotation editor (removed in Grafana 12) with a React editor. Existing dashboard annotations are migrated automatically.
- Rebuilt the frontend with `@grafana/create-plugin` and current `@grafana/ui` components.
- Updated the backend to the current Grafana plugin SDK and a modern Go toolchain.
- Caches are now scoped to each datasource instance, so data can no longer be shared between different servers or credentials.
- All values interpolated into Octopus API requests are validated against strict allow-lists.
- Response bodies are no longer written to the Grafana server log.
- Query failures now surface as panel errors instead of silently empty results.
- The space, project, environment, channel, tenant and task state filters are now dropdowns populated from Octopus, while still accepting typed values such as template variables.
- Queries that match no deployments attach a panel notice stating when the most recent matching deployment happened, instead of showing a bare "No data".
- The datasource is now declared alerting-capable, so its queries can back Grafana alert rules.
- "Save & test" warns when the Octopus server URL uses plain HTTP, since the API key is sent unencrypted.

## 1.x

Initial early-access releases targeting Grafana 7.
