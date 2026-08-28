import React, { ChangeEvent } from 'react';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { InlineField, Input, SecretInput } from '@grafana/ui';

import { MyDataSourceOptions, MySecureJsonData } from '../types';

interface Props extends DataSourcePluginOptionsEditorProps<MyDataSourceOptions, MySecureJsonData> {}

export function ConfigEditor(props: Props) {
  const { onOptionsChange, options } = props;
  const { jsonData, secureJsonFields, secureJsonData } = options;

  const onServerChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        server: event.target.value,
      },
    });
  };

  const onCacheDurationChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        cacheDuration: event.target.value,
      },
    });
  };

  // Secure field (only sent to the backend)
  const onAPIKeyChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      secureJsonData: {
        apiKey: event.target.value,
      },
    });
  };

  const onResetAPIKey = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        apiKey: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        apiKey: '',
      },
    });
  };

  return (
    <>
      <InlineField
        label="Server"
        labelWidth={20}
        interactive
        tooltip="The URL of the Octopus server, e.g. https://myinstance.octopus.app"
      >
        <Input
          id="config-editor-server"
          onChange={onServerChange}
          value={jsonData.server || ''}
          placeholder="https://octopusserver"
          width={40}
        />
      </InlineField>
      <InlineField
        label="API Key"
        labelWidth={20}
        interactive
        tooltip="An Octopus API key with permission to view deployments, environments, tenants, processes, projects and releases"
      >
        <SecretInput
          required
          id="config-editor-api-key"
          isConfigured={!!secureJsonFields.apiKey}
          value={secureJsonData?.apiKey || ''}
          placeholder="API-xxxxxxxxxx"
          width={40}
          onReset={onResetAPIKey}
          onChange={onAPIKeyChange}
        />
      </InlineField>
      <InlineField
        label="Cache Duration"
        labelWidth={20}
        interactive
        tooltip="How long entity lookups (projects, environments, ...) are cached, e.g. 1m or 30s. Leave blank to disable caching."
      >
        <Input
          id="config-editor-cache-duration"
          onChange={onCacheDurationChange}
          value={jsonData.cacheDuration || ''}
          placeholder="1m"
          width={40}
        />
      </InlineField>
    </>
  );
}
