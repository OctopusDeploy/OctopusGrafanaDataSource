import defaults from 'lodash/defaults';

import React, { ChangeEvent, PureComponent } from 'react';
import {InlineFormLabel, LegacyForms, Switch} from '@grafana/ui';
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

  onSuccessFieldSwitchChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, successField: event.target.checked });
  };

  onFailureFieldSwitchChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, failureField: event.target.checked });
  };

  onCancelledFieldSwitchChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, cancelledField: event.target.checked });
  };

  onTimedOutFieldSwitchChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, timedOutField: event.target.checked });
  };

  onTotalDurationFieldSwitchChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, totalDurationField: event.target.checked });
  };

  onAverageDurationFieldSwitchChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, averageDurationField: event.target.checked });
  };

  render() {
    const query = defaults(this.props.query, defaultQuery);
    const { projectName, environmentName, channelName, tenantName, releaseVersion, format, successField, failureField, cancelledField, timedOutField, totalDurationField, averageDurationField } = query;
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
        <InlineFormLabel>Success field</InlineFormLabel>
        <Switch
          css={null}
          value={successField == null ? true : successField}
          onChange={this.onSuccessFieldSwitchChange}
        />
        <InlineFormLabel>Failure field</InlineFormLabel>
        <Switch
          css={null}
          value={failureField == null ? true : failureField}
          onChange={this.onFailureFieldSwitchChange}
        />
        <InlineFormLabel>Cancelled field</InlineFormLabel>
        <Switch
          css={null}
          value={cancelledField == null ? true : cancelledField}
          onChange={this.onCancelledFieldSwitchChange}
        />

        <InlineFormLabel>Timed Out field</InlineFormLabel>
        <Switch
          css={null}
          value={timedOutField == null ? true : timedOutField}
          onChange={this.onTimedOutFieldSwitchChange}
        />

        <InlineFormLabel>Total Duration Field</InlineFormLabel>
        <Switch
          css={null}
          value={totalDurationField == null ? true : totalDurationField}
          onChange={this.onTotalDurationFieldSwitchChange}
        />

        <InlineFormLabel>Average Duration Field</InlineFormLabel>
        <Switch
          css={null}
          value={averageDurationField == null ? true : averageDurationField}
          onChange={this.onAverageDurationFieldSwitchChange}
        />
      </div>
    );
  }
}
