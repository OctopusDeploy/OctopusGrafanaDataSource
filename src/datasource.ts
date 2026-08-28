import {
  AnnotationQuery,
  AnnotationSupport,
  CoreApp,
  CustomVariableSupport,
  DataQueryRequest,
  DataQueryResponse,
  DataSourceInstanceSettings,
  MetricFindValue,
  ScopedVars,
} from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';
import { from, map, Observable } from 'rxjs';

import { AnnotationQueryEditor } from './components/AnnotationQueryEditor';
import { VariableQueryEditor } from './components/VariableQueryEditor';
import { ANNOTATION_FORMAT_PREFIX, DEFAULT_QUERY, MyDataSourceOptions, MyQuery, MyVariableQuery } from './types';

export class DataSource extends DataSourceWithBackend<MyQuery, MyDataSourceOptions> {
  annotations: AnnotationSupport<MyQuery> = {
    QueryEditor: AnnotationQueryEditor,
    prepareAnnotation: (json: AnnotationQuery<MyQuery> & Partial<MyQuery>): AnnotationQuery<MyQuery> => {
      if (json.target?.format?.startsWith(ANNOTATION_FORMAT_PREFIX)) {
        return json;
      }

      // Migrate annotations saved by the pre-Grafana 11 (Angular) editor,
      // which stored the query fields on the annotation itself and used the
      // unprefixed format names "deploymentreport" and "deployments".
      const legacyFormat = json.target?.format ?? json.format ?? 'deploymentreport';
      const target: MyQuery = {
        refId: json.target?.refId ?? 'Anno',
        format: legacyFormat.startsWith(ANNOTATION_FORMAT_PREFIX)
          ? legacyFormat
          : `${ANNOTATION_FORMAT_PREFIX}${legacyFormat}`,
        spaceName: json.target?.spaceName ?? json.spaceName ?? '',
        projectName: json.target?.projectName ?? json.projectName ?? '',
        environmentName: json.target?.environmentName ?? json.environmentName ?? '',
      };
      return { ...json, target };
    },
    prepareQuery: (anno: AnnotationQuery<MyQuery>): MyQuery | undefined => {
      if (!anno.target?.format) {
        return undefined;
      }
      return { ...anno.target, refId: anno.target.refId ?? 'Anno' };
    },
  };

  constructor(instanceSettings: DataSourceInstanceSettings<MyDataSourceOptions>) {
    super(instanceSettings);
    this.variables = new OctopusVariableSupport(this);
  }

  getDefaultQuery(_: CoreApp): Partial<MyQuery> {
    return DEFAULT_QUERY;
  }

  applyTemplateVariables(query: MyQuery, scopedVars: ScopedVars): MyQuery {
    const templateSrv = getTemplateSrv();

    return {
      ...query,
      spaceName: query.spaceName ? templateSrv.replace(query.spaceName, scopedVars) : '',
      projectName: query.projectName ? templateSrv.replace(query.projectName, scopedVars) : '',
      tenantName: query.tenantName ? templateSrv.replace(query.tenantName, scopedVars) : '',
      environmentName: query.environmentName ? templateSrv.replace(query.environmentName, scopedVars) : '',
      channelName: query.channelName ? templateSrv.replace(query.channelName, scopedVars) : '',
      releaseVersion: query.releaseVersion ? templateSrv.replace(query.releaseVersion, scopedVars) : '',
      taskState: query.taskState ? templateSrv.replace(query.taskState, scopedVars) : '',
    };
  }

  /**
   * Returns the names of the requested entity type, scoped to a space unless
   * the entity type is spaces itself. An empty space name selects the default
   * space, which the backend maps from a single space character.
   */
  async getEntityNames(entityType: string, spaceName?: string): Promise<string[]> {
    const spaces = await this.getResource<Record<string, string>>('spaces/nameid');

    if (entityType === 'spaces') {
      return Object.keys(spaces)
        .filter((name) => name.trim() !== '')
        .sort();
    }

    const resolvedSpaceName = getTemplateSrv().replace(spaceName || '') || ' ';
    const spaceId = spaces[resolvedSpaceName];
    if (!spaceId) {
      return [];
    }

    const entities = await this.getResource<Record<string, string>>(
      `${encodeURIComponent(spaceId)}/nameid/${encodeURIComponent(entityType)}`
    );
    return Object.keys(entities)
      .filter((name) => name.trim() !== '')
      .sort();
  }

  /**
   * Variable query action.
   */
  async metricFindQuery(query: MyVariableQuery): Promise<MetricFindValue[]> {
    if (!query?.entityName) {
      return [];
    }

    const names = await this.getEntityNames(query.entityName, query.spaceName);
    return names.map((text) => ({ text }));
  }
}

class OctopusVariableSupport extends CustomVariableSupport<DataSource, MyVariableQuery> {
  editor = VariableQueryEditor;

  constructor(private datasource: DataSource) {
    super();
  }

  query(request: DataQueryRequest<MyVariableQuery>): Observable<DataQueryResponse> {
    return from(this.datasource.metricFindQuery(request.targets[0])).pipe(map((data) => ({ data })));
  }
}
