import { test, expect } from '@grafana/plugin-e2e';

test('smoke: should render query editor', async ({ panelEditPage, readProvisionedDataSource }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  await expect(panelEditPage.getQueryEditorRow('A').getByText('Result Format')).toBeVisible();
  await expect(panelEditPage.getQueryEditorRow('A').getByRole('combobox', { name: 'Space Name Filter' })).toBeVisible();
});

test('should show the deployment filters for the time series format only', async ({
  panelEditPage,
  readProvisionedDataSource,
  selectors,
}) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  const row = panelEditPage.getQueryEditorRow('A');

  await expect(row.getByRole('combobox', { name: 'Project Name Filter' })).toBeVisible();
  await expect(row.getByText('Return Success Field')).toBeVisible();

  await row.getByRole('combobox', { name: 'Result Format' }).click();
  await panelEditPage.getByGrafanaSelector(selectors.components.Select.option).getByText('environments table').click();

  await expect(row.getByRole('combobox', { name: 'Project Name Filter' })).not.toBeVisible();
  await expect(row.getByText('Return Success Field')).not.toBeVisible();
});

test('space filter should list the spaces from Octopus', async ({
  panelEditPage,
  readProvisionedDataSource,
  selectors,
}) => {
  // Requires a reachable Octopus server; pass its URL and API key via the
  // OCTOPUS_SERVER and OCTOPUS_API_KEY environment variables.
  test.skip(!process.env.OCTOPUS_SERVER, 'OCTOPUS_SERVER is not set');

  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  await panelEditPage.getQueryEditorRow('A').getByRole('combobox', { name: 'Space Name Filter' }).click();
  await expect(
    panelEditPage.getByGrafanaSelector(selectors.components.Select.option).getByText('Default')
  ).toBeVisible();
});
