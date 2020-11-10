import defaults from 'lodash/defaults';

import React, { ChangeEvent, PureComponent } from 'react';
import { LegacyForms } from '@grafana/ui';
import {QueryEditorProps, SelectableValue} from '@grafana/data';
import { DataSource } from './DataSource';
import { defaultQuery, MyDataSourceOptions, MyQuery } from './types';

const { FormField, Select } = LegacyForms;

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export class QueryEditor extends PureComponent<Props> {
  onFormatTextChange = (value: SelectableValue<string>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, format: value.value });
  };

  onProjectNameTextChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, projectName: event.target.value });
  };

  onEnvironmentNameTextChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, environmentName: event.target.value });
  };

  onTenantNameTextChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, tenantName: event.target.value });
  };

  onChannelNameTextChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, channelName: event.target.value });
  };

  onReleaseVersionTextChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, releaseVersion: event.target.value });
  };

  onConstantChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query });
    // executes the query
    onRunQuery();
  };

  render() {
    const query = defaults(this.props.query, defaultQuery);
    const { projectName, environmentName, channelName, tenantName, releaseVersion, format } = query;
    const formatOptions = [{value: "timeseries", label: "time series"}, {value: "table", label: "table"}];

    return (
      <div className="gf-form" style={{flexDirection: "column"}}>
        <Select
          value={formatOptions.find(f => f.value == format) || formatOptions.find(f => f.value == "timeseries")}
          options={formatOptions}
          onChange={this.onFormatTextChange}
        />
        <FormField
          labelWidth={8}
          value={projectName || ''}
          onChange={this.onProjectNameTextChange}
          label="Project Name"
        />
        <FormField
          labelWidth={8}
          value={environmentName || ''}
          onChange={this.onEnvironmentNameTextChange}
          label="Environment Name"
        />
        <FormField
          labelWidth={8}
          value={channelName || ''}
          onChange={this.onChannelNameTextChange}
          label="Channel Name"
        />
        <FormField
          labelWidth={8}
          value={tenantName || ''}
          onChange={this.onTenantNameTextChange}
          label="Tenant Name"
        />
        <FormField
          labelWidth={8}
          value={releaseVersion || ''}
          onChange={this.onReleaseVersionTextChange}
          label="Release Version"
        />
      </div>
    );
  }
}
