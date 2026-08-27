import { test, expect } from '@grafana/plugin-e2e';

test('variable query should return space names', async ({
  variableEditPage,
  readProvisionedDataSource,
  page,
  selectors,
}) => {
  // Requires a reachable Octopus server; pass its URL and API key via the
  // OCTOPUS_SERVER and OCTOPUS_API_KEY environment variables.
  test.skip(!process.env.OCTOPUS_SERVER, 'OCTOPUS_SERVER is not set');

  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await variableEditPage.datasource.set(ds.name);
  await page.getByRole('combobox', { name: 'Entity Type' }).click();
  await variableEditPage.getByGrafanaSelector(selectors.components.Select.option).getByText('spaces').click();
  await variableEditPage.runQuery();
  await expect(variableEditPage).toDisplayPreviews(['Default']);
});
