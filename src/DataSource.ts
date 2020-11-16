import {DataSourceInstanceSettings, MetricFindValue} from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';
import { MyDataSourceOptions, MyQuery } from './types';
import {MyVariableQuery} from "./VariableQueryEditor";

export class DataSource extends DataSourceWithBackend<MyQuery, MyDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<MyDataSourceOptions>) {
    super(instanceSettings);
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
    const url = await this.getUrl(query, options);
    return fetch(url)
      .then(response => response.json())
      .then(data => {
        if (data) {
          return data.map((text: string) => ({text}))
        }
        return [];
      });
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

    return `/api/datasources/${options.variable.datasource}/resources/${spaces[query.spaceName]}/${query.entityName}`;
  }
}
