import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

export interface MyQuery extends DataQuery {
  spaceName?: string;
  projectName?: string;
  tenantName?: string;
  environmentName?: string;
  channelName?: string;
  releaseVersion?: string;
  taskState?: string;
  format?: string;
  successField?: boolean;
  failureField?: boolean;
  timedOutField?: boolean;
  cancelledField?: boolean;
  totalDurationField?: boolean;
  averageDurationField?: boolean;
  totalTimeToRecoveryField?: boolean;
  averageTimeToRecoveryField?: boolean;
  totalCycleTimeField?: boolean;
  averageCycleTimeField?: boolean;
}

export const DEFAULT_QUERY: Partial<MyQuery> = {
  format: 'timeseries',
  successField: true,
  failureField: true,
  timedOutField: true,
  cancelledField: true,
  totalDurationField: true,
  averageDurationField: true,
  totalTimeToRecoveryField: true,
  averageTimeToRecoveryField: true,
};

/**
 * Annotation queries are regular queries with one of these formats. Legacy
 * annotations stored the unprefixed format names, which are migrated in
 * prepareAnnotation.
 */
export const ANNOTATION_FORMAT_PREFIX = 'annotation-';
export const ANNOTATION_FORMAT_DEPLOYMENT_REPORT = 'annotation-deploymentreport';
export const ANNOTATION_FORMAT_DEPLOYMENTS = 'annotation-deployments';

export interface MyVariableQuery extends DataQuery {
  spaceName: string;
  entityName: string;
}

/**
 * These are options configured for each DataSource instance
 */
export interface MyDataSourceOptions extends DataSourceJsonData {
  server?: string;
  cacheDuration?: string;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface MySecureJsonData {
  apiKey?: string;
}
