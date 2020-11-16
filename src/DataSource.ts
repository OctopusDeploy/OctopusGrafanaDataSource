import {DataSourceInstanceSettings, MetricFindValue} from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';
import { getTemplateSrv } from '@grafana/runtime';
import { MyDataSourceOptions, MyQuery } from './types';
import {MyVariableQuery} from "./VariableQueryEditor";

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
  async metricFindQuery(query: MyVariableQuery, options?: any) : Promise<MetricFindValue[]> {
    /**
     * If query or datasource not specified
     */
    if (!query || !options.variable.datasource || !query.entityName || (query.entityName != "spaces" && !query.spaceName)) {
      return Promise.resolve([]);
    }

    /**
     * Run Query
     */
    return this.getUrl(query, options)
      .then(url => fetch(url))
      .then(response => response.json())
      .then(data => {
        if (data) {
          return data.map((text: string) => ({text}))
        }
        return [];
      })
      .catch(() => [])
  }

  async getUrl(query: MyVariableQuery, options?: any) {
    if (query.entityName == "spaces") {
      return `/api/datasources/${options.variable.datasource}/resources/spaces`;
    }

    /**
     * Get space names mapped to IDs
     */
    const spaces = await fetch(`/api/datasources/${options.variable.datasource}/resources/spacesMapping`)
      .then(response => response.json());

    const spaceName = getTemplateSrv().replace(query.spaceName, options.scopedVars)

    if (spaces[spaceName]) {
      return `/api/datasources/${options.variable.datasource}/resources/${spaces[spaceName]}/${query.entityName}`;
    }

    throw "Space could not be found"
  }
}
