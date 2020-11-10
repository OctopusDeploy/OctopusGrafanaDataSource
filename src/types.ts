import { DataQuery, DataSourceJsonData } from '@grafana/data';

export interface MyQuery extends DataQuery {
  projectName?: string;
  tenantName?: string;
  environmentName?: string;
  channelName?: string;
  releaseVersion?: string;
}

export const defaultQuery: Partial<MyQuery> = {

};

/**
 * These are options configured for each DataSource instance
 */
export interface MyDataSourceOptions extends DataSourceJsonData {
  server?: string;
  spaceId?: string;
  bucketDuration?: string;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface MySecureJsonData {
  apiKey?: string;
}
