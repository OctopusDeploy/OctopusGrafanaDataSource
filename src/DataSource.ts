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
    if (!query || !options.variable.datasource || !query.spaceName || !query.entityName) {
      return Promise.resolve([]);
    }

    /**
     * Get space names mapped to IDs
     */
    const spacesUrl = `/api/datasources/${options.variable.datasource}/resources/spaces`;
    const spaces = await fetch(spacesUrl)
      .then(response => response.json());

    /**
     * Run Query
     */
    const url = `/api/datasources/${options.variable.datasource}/resources/${spaces[query.spaceName]}/${query.entityName}`;
    return fetch(url)
      .then(response => response.json())
      .then(data => {
        if (data) {
          return data.map((text: string) => ({text}))
        }
        return [];
      });
  }
}
