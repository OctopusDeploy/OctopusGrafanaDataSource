import { AnnotationEvent, AnnotationQueryRequest, DataSourceInstanceSettings, MetricFindValue } from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';
import { getTemplateSrv } from '@grafana/runtime';
import { MyDataSourceOptions, MyQuery } from './types';
import { MyVariableQuery } from './VariableQueryEditor';
import _ from 'lodash';

export class DataSource extends DataSourceWithBackend<MyQuery, MyDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<MyDataSourceOptions>) {
    super(instanceSettings);
  }

  applyTemplateVariables(query: MyQuery) {
    const templateSrv = getTemplateSrv();

    return {
      ...query,
      spaceName: query.spaceName ? templateSrv.replace(query.spaceName) : '',
      projectName: query.projectName ? templateSrv.replace(query.projectName) : '',
      tenantName: query.tenantName ? templateSrv.replace(query.tenantName) : '',
      environmentName: query.environmentName ? templateSrv.replace(query.environmentName) : '',
      channelName: query.channelName ? templateSrv.replace(query.channelName) : '',
      releaseVersion: query.releaseVersion ? templateSrv.replace(query.releaseVersion) : '',
      taskState: query.taskState ? templateSrv.replace(query.taskState) : '',
    };
  }

  /**
   * Variable query action.
   */
  async metricFindQuery(query: MyVariableQuery, options?: any): Promise<MetricFindValue[]> {
    /**
     * If query or datasource not specified
     */
    if (
      !query ||
      !options.variable.datasource ||
      !query.entityName ||
      (query.entityName !== 'spaces' && !query.spaceName)
    ) {
      return Promise.resolve([]);
    }

    /**
     * Run Query
     */
    return this.getUrl(query.entityName, query.spaceName, options.variable.datasource)
      .then(url => fetch(url))
      .then(response => response.json())
      .then(data => {
        if (data) {
          return Object.keys(data).map((text: string) => ({ text }));
        }
        return [];
      })
      .catch(() => []);
  }

  async getUrl(entityName: string, spaceName: string, datasource: string) {
    if (entityName === 'spaces') {
      return `/api/datasources/${datasource}/resources/spaces`;
    }

    /**
     * Get space names mapped to IDs
     */
    const spaces = await fetch(`/api/datasources/${datasource}/resources/spaces`).then(response => response.json());

    const spaceNameFixed = getTemplateSrv().replace(spaceName);

    if (spaces[spaceNameFixed]) {
      return `/api/datasources/${datasource}/resources/${spaces[spaceNameFixed]}/${entityName}`;
    }

    throw 'Space could not be found';
  }

  /**
   * This method returns the deployment history to be overlayed on a graph as an annotation. It
   * called the backend plugin resource endpoint to get the deployments history, and converts it into
   * an array of annotation events.
   */
  async annotationQuery(options: AnnotationQueryRequest<MyQuery>): Promise<AnnotationEvent[]> {
    const query: MyQuery = options.annotation;
    const datasource = options.annotation.datasource;

    const spaceId = await this.getSpaceId(query.spaceName || '', datasource);
    const environmentId = await this.getEntityId(
      query.spaceName || '',
      'environments',
      query.environmentName || '',
      datasource
    );
    const projectId = await this.getEntityId(query.spaceName || '', 'projects', query.projectName || '', datasource);
    const from = options.range.from.format('YYYY-MM-DD HH:mm:ss');
    const to = options.range.to.format('YYYY-MM-DD HH:mm:ss');

    const url =
      `/api/datasources/${datasource}/resources/${spaceId}/reporting/deployments` +
      '?environmentId=' +
      encodeURI(environmentId) +
      '&projectId=' +
      encodeURI(projectId) +
      '&fromCompletedTime=' +
      encodeURI(from) +
      '&toCompletedTime=' +
      encodeURI(to);

    return fetch(url)
      .then(response => response.json())
      .then(data =>
        data.Deployments
          ? data.Deployments.map((d: any) => ({
              time: Date.parse(d.StartTimeParsed),
              timeEnd: Date.parse(d.CompletedTimeParsed),
              isRegion: true,
              text: d.DeploymentId,
              tags: [
                'Project: ' + d.ProjectName,
                'Tenant: ' + d.TenantName,
                'Channel: ' + d.ChannelName,
                'Environment: ' + d.EnvironmentName,
                'Version: ' + d.ReleaseVersion,
              ],
            }))
          : []
      );
  }

  /**
   * Convert an entity name into an ID.
   * @param spaceName The name of the space.
   * @param entityType The type of the entity being converted.
   * @param entityName The name of the entity being converted.
   * @param datasource The name of the datasource.
   * @return The ID of the entity.
   */
  async getEntityId(spaceName: string, entityType: string, entityName: string, datasource: string): Promise<string> {
    const entityNameFixed = getTemplateSrv().replace(entityName);
    const url = await this.getUrl(entityType, spaceName, datasource);
    const entities = await fetch(url).then(response => response.json());
    return entities[entityNameFixed] || '';
  }

  /**
   * Convert a space name into an ID.
   * @param spaceName The name of the space.
   * @param datasource The name of the datasource.
   * @return The ID of the space.
   */
  async getSpaceId(spaceName: string, datasource: string): Promise<string> {
    const entityNameFixed = getTemplateSrv().replace(spaceName);
    const url = await this.getUrl('spaces', entityNameFixed, datasource);
    const entities = await fetch(url).then(response => response.json());
    return entities[entityNameFixed] || '';
  }
}
