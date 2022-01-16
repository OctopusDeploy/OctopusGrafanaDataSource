FROM grafana/grafana:8.3.3
RUN mkdir /var/lib/grafana/plugins/octopus
RUN mkdir /var/lib/grafana/plugins/octopus/img
COPY ./* /var/lib/grafana/plugins/octopus/
COPY ./img/* /var/lib/grafana/plugins/octopus/img/
ENV GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=octopus-deploy-xmlfeed
