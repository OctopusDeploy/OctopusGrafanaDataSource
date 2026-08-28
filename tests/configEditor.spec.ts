import { test, expect } from '@grafana/plugin-e2e';

test('smoke: should render config editor', async ({ createDataSourceConfigPage, readProvisionedDataSource, page }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await createDataSourceConfigPage({ type: ds.type });
  await expect(page.getByLabel('Server')).toBeVisible();
  await expect(page.getByLabel('API Key')).toBeVisible();
  await expect(page.getByLabel('Cache Duration')).toBeVisible();
});

test('"Save & test" should be successful when configuration is valid', async ({
  createDataSourceConfigPage,
  readProvisionedDataSource,
  page,
}) => {
  // Requires a reachable Octopus server; pass its URL and API key via the
  // OCTOPUS_SERVER and OCTOPUS_API_KEY environment variables.
  test.skip(!process.env.OCTOPUS_SERVER, 'OCTOPUS_SERVER is not set');

  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  const configPage = await createDataSourceConfigPage({ type: ds.type });
  await page.getByRole('textbox', { name: 'Server' }).fill(process.env.OCTOPUS_SERVER ?? '');
  await page.getByRole('textbox', { name: 'API Key' }).fill(process.env.OCTOPUS_API_KEY ?? '');
  await expect(configPage.saveAndTest()).toBeOK();
});

test('"Save & test" should fail when the server is unreachable', async ({
  createDataSourceConfigPage,
  readProvisionedDataSource,
  page,
}) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  const configPage = await createDataSourceConfigPage({ type: ds.type });
  await page.getByRole('textbox', { name: 'Server' }).fill('http://localhost:1');
  await page.getByRole('textbox', { name: 'API Key' }).fill('API-INVALID');
  await expect(configPage.saveAndTest()).not.toBeOK();
});

test('"Save & test" should fail when the server URL is not http or https', async ({
  createDataSourceConfigPage,
  readProvisionedDataSource,
  page,
}) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  const configPage = await createDataSourceConfigPage({ type: ds.type });
  await page.getByRole('textbox', { name: 'Server' }).fill('ftp://example');
  await page.getByRole('textbox', { name: 'API Key' }).fill('API-INVALID');
  await expect(configPage.saveAndTest()).not.toBeOK();
});
