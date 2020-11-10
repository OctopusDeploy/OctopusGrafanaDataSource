FROM grafana/grafana:7.0.0
RUN mkdir /var/lib/grafana/plugins/octopus
RUN mkdir /var/lib/grafana/plugins/octopus/img
COPY dist/* /var/lib/grafana/plugins/octopus/
COPY dist/img/* /var/lib/grafana/plugins/octopus/img/
ENV GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=octopus-deploy-xmlfeed
