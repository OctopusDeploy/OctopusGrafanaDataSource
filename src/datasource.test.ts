import { AnnotationQuery, DataSourceInstanceSettings } from '@grafana/data';

import { DataSource } from './datasource';
import { MyDataSourceOptions, MyQuery } from './types';

jest.mock('@grafana/runtime', () => ({
  ...jest.requireActual('@grafana/runtime'),
  getTemplateSrv: () => ({
    replace: (value?: string) => value ?? '',
  }),
}));

function createDataSource(): DataSource {
  const settings = {
    id: 1,
    uid: 'octopus',
    jsonData: {},
  } as DataSourceInstanceSettings<MyDataSourceOptions>;
  return new DataSource(settings);
}

describe('prepareAnnotation', () => {
  it('migrates legacy Angular annotations to a query target', () => {
    const ds = createDataSource();
    const legacy = {
      name: 'Deployments',
      enable: true,
      iconColor: 'red',
      spaceName: 'Default',
      projectName: 'Web',
      environmentName: 'Dev',
      format: 'deploymentreport',
    } as unknown as AnnotationQuery<MyQuery>;

    const migrated = ds.annotations.prepareAnnotation!(legacy);

    expect(migrated.target).toEqual({
      refId: 'Anno',
      format: 'annotation-deploymentreport',
      spaceName: 'Default',
      projectName: 'Web',
      environmentName: 'Dev',
    });
  });

  it('defaults legacy annotations without a format to deployment reports', () => {
    const ds = createDataSource();
    const legacy = { name: 'Deployments', enable: true } as unknown as AnnotationQuery<MyQuery>;

    const migrated = ds.annotations.prepareAnnotation!(legacy);

    expect(migrated.target?.format).toBe('annotation-deploymentreport');
  });

  it('leaves already migrated annotations untouched', () => {
    const ds = createDataSource();
    const annotation = {
      name: 'Deployments',
      enable: true,
      target: {
        refId: 'Anno',
        format: 'annotation-deployments',
        spaceName: 'Default',
        projectName: '',
        environmentName: '',
      },
    } as AnnotationQuery<MyQuery>;

    expect(ds.annotations.prepareAnnotation!(annotation)).toBe(annotation);
  });
});

describe('prepareQuery', () => {
  it('returns the annotation target as the query', () => {
    const ds = createDataSource();
    const annotation = {
      name: 'Deployments',
      enable: true,
      target: { refId: 'Anno', format: 'annotation-deployments', spaceName: 'Default' },
    } as AnnotationQuery<MyQuery>;

    expect(ds.annotations.prepareQuery!(annotation)).toEqual(annotation.target);
  });

  it('returns undefined when there is nothing to query', () => {
    const ds = createDataSource();
    const annotation = { name: 'Deployments', enable: true } as AnnotationQuery<MyQuery>;

    expect(ds.annotations.prepareQuery!(annotation)).toBeUndefined();
  });
});

describe('getEntityNames', () => {
  it('sorts names and hides the default space alias', async () => {
    const ds = createDataSource();
    ds.getResource = jest.fn().mockResolvedValue({ Zeta: 'Spaces-2', Alpha: 'Spaces-1', ' ': 'Spaces-1' });

    await expect(ds.getEntityNames('spaces')).resolves.toEqual(['Alpha', 'Zeta']);
  });

  it('scopes entity lookups to the resolved space', async () => {
    const ds = createDataSource();
    ds.getResource = jest
      .fn()
      .mockResolvedValueOnce({ Default: 'Spaces-1', ' ': 'Spaces-1' })
      .mockResolvedValueOnce({ Web: 'Projects-1', Api: 'Projects-2' });

    await expect(ds.getEntityNames('projects', 'Default')).resolves.toEqual(['Api', 'Web']);
    expect(ds.getResource).toHaveBeenLastCalledWith('Spaces-1/nameid/projects');
  });
});

describe('metricFindQuery', () => {
  it('lists space names', async () => {
    const ds = createDataSource();
    ds.getResource = jest.fn().mockResolvedValue({ Default: 'Spaces-1', ' ': 'Spaces-1' });

    const values = await ds.metricFindQuery({ refId: 'A', entityName: 'spaces', spaceName: '' });

    expect(values).toEqual([{ text: 'Default' }]);
    expect(ds.getResource).toHaveBeenCalledWith('spaces/nameid');
  });

  it('lists entities within the requested space', async () => {
    const ds = createDataSource();
    ds.getResource = jest
      .fn()
      .mockResolvedValueOnce({ Default: 'Spaces-1', ' ': 'Spaces-1' })
      .mockResolvedValueOnce({ Web: 'Projects-1' });

    const values = await ds.metricFindQuery({ refId: 'A', entityName: 'projects', spaceName: 'Default' });

    expect(values).toEqual([{ text: 'Web' }]);
    expect(ds.getResource).toHaveBeenLastCalledWith('Spaces-1/nameid/projects');
  });

  it('uses the default space when no space name is given', async () => {
    const ds = createDataSource();
    ds.getResource = jest
      .fn()
      .mockResolvedValueOnce({ Default: 'Spaces-1', ' ': 'Spaces-1' })
      .mockResolvedValueOnce({ Dev: 'Environments-1' });

    const values = await ds.metricFindQuery({ refId: 'A', entityName: 'environments', spaceName: '' });

    expect(values).toEqual([{ text: 'Dev' }]);
    expect(ds.getResource).toHaveBeenLastCalledWith('Spaces-1/nameid/environments');
  });

  it('returns nothing for an unknown space', async () => {
    const ds = createDataSource();
    ds.getResource = jest.fn().mockResolvedValue({ Default: 'Spaces-1' });

    const values = await ds.metricFindQuery({ refId: 'A', entityName: 'projects', spaceName: 'Nope' });

    expect(values).toEqual([]);
  });
});
