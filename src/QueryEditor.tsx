import defaults from 'lodash/defaults';

import React, { ChangeEvent, PureComponent } from 'react';
import { InlineFormLabel, LegacyForms, Switch } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataSource } from './DataSource';
import { defaultQuery, MyDataSourceOptions, MyQuery } from './types';

const { FormField, Select } = LegacyForms;

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export class QueryEditor extends PureComponent<Props> {
  onFormatTextChange = (value: SelectableValue<string>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, format: value.value });
  };

  onSpaceNameTextChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, spaceName: event.target.value });
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

  onTaskSTateTextChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, taskState: event.target.value });
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

  onTotalTimeToRecoveryFieldSwitchChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, totalTimeToRecoveryField: event.target.checked });
  };

  onAverageTimeToRecoveryFieldSwitchChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, averageTimeToRecoveryField: event.target.checked });
  };

  onTotalCycleTimeFieldSwitchChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, totalCycleTimeField: event.target.checked });
  };

  onAverageCycleTimeFieldSwitchChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, averageCycleTimeField: event.target.checked });
  };

  render() {
    const query = defaults(this.props.query, defaultQuery);
    const {
      spaceName,
      projectName,
      environmentName,
      channelName,
      tenantName,
      releaseVersion,
      taskState,
      format,
      successField,
      failureField,
      cancelledField,
      timedOutField,
      totalDurationField,
      averageDurationField,
      totalTimeToRecoveryField,
      averageTimeToRecoveryField,
      totalCycleTimeField,
      averageCycleTimeField,
    } = query;
    const formatOptions = [
      { value: 'timeseries', label: 'deployments time series' },
      { value: 'table', label: 'deployments table' },
      { value: 'accounts', label: 'accounts table' },
      { value: 'actiontemplates', label: 'action templates table' },
      { value: 'certificates', label: 'certificates table' },
      { value: 'feeds', label: 'feeds table' },
      { value: 'libraryvariablesets', label: 'library variable sets table' },
      { value: 'machinepolicies', label: 'machine policies table' },
      { value: 'machineroles', label: 'machine roles table' },
      { value: 'machines', label: 'targets table' },
      { value: 'octopusservernodes', label: 'octopus server nodes table' },
      { value: 'permissions', label: 'permissions table' },
      { value: 'projectgroups', label: 'project groups table' },
      { value: 'proxies', label: 'proxies table' },
      { value: 'releases', label: 'releases table' },
      { value: 'runbooks', label: 'runbooks table' },
      { value: 'spaces', label: 'spaces table' },
      { value: 'subscriptions', label: 'subscriptions table' },
      { value: 'tagsets', label: 'tag sets table' },
      { value: 'teams', label: 'teams table' },
      { value: 'tenantvariables', label: 'tenant variables table' },
      { value: 'roles', label: 'roles table' },
      { value: 'users', label: 'users table' },
      { value: 'variables', label: 'variable sets table' },
      { value: 'workerpools', label: 'worker pools table' },
      { value: 'workers', label: 'workers table' },
      { value: 'environments', label: 'environments table' },
      { value: 'tenants', label: 'tenants table' },
      { value: 'channels', label: 'channels table' },
      { value: 'projects', label: 'projects table' },
    ];

    return (
      <div className="gf-form" style={{ flexDirection: 'column' }}>
        <div style={{ alignContent: 'flex-start', flexWrap: 'wrap', display: 'flex', flexDirection: 'row' }}>
          <InlineFormLabel width={20}>Result Format</InlineFormLabel>
          <Select
            value={formatOptions.find(f => f.value === format) || formatOptions.find(f => f.value === 'timeseries')}
            options={formatOptions}
            onChange={this.onFormatTextChange}
          />
        </div>
        <FormField
          labelWidth={20}
          value={spaceName || ''}
          onChange={this.onSpaceNameTextChange}
          label="Space Name Filter"
        />
        {(format === 'timeseries' || format === 'table') && (
          <div>
            <FormField
              labelWidth={20}
              value={projectName || ''}
              onChange={this.onProjectNameTextChange}
              label="Project Name Filter"
            />
            <FormField
              labelWidth={20}
              value={environmentName || ''}
              onChange={this.onEnvironmentNameTextChange}
              label="Environment Name Filter"
            />
            <FormField
              labelWidth={20}
              value={channelName || ''}
              onChange={this.onChannelNameTextChange}
              label="Channel Name Filter"
            />
            <FormField
              labelWidth={20}
              value={tenantName || ''}
              onChange={this.onTenantNameTextChange}
              label="Tenant Name Filter"
            />
            <FormField
              labelWidth={20}
              value={releaseVersion || ''}
              onChange={this.onReleaseVersionTextChange}
              label="Release Version Filter"
            />
            <FormField
              labelWidth={20}
              value={taskState || ''}
              onChange={this.onTaskSTateTextChange}
              label="Task State Filter"
            />
            {format === 'timeseries' && (
              <div>
                <div style={{ alignContent: 'flex-start', flexWrap: 'wrap', display: 'flex', flexDirection: 'row' }}>
                  <InlineFormLabel width={20}>Return Success Field</InlineFormLabel>
                  <Switch
                    css="css"
                    value={successField === null ? true : successField}
                    onChange={this.onSuccessFieldSwitchChange}
                  />
                </div>
                <div style={{ alignContent: 'flex-start', flexWrap: 'wrap', display: 'flex', flexDirection: 'row' }}>
                  <InlineFormLabel width={20}>Return Failure Field</InlineFormLabel>
                  <Switch
                    css="css"
                    value={failureField === null ? true : failureField}
                    onChange={this.onFailureFieldSwitchChange}
                  />
                </div>
                <div style={{ alignContent: 'flex-start', flexWrap: 'wrap', display: 'flex', flexDirection: 'row' }}>
                  <InlineFormLabel width={20}>Return Cancelled Field</InlineFormLabel>
                  <Switch
                    css="css"
                    value={cancelledField === null ? true : cancelledField}
                    onChange={this.onCancelledFieldSwitchChange}
                  />
                </div>
                <div style={{ alignContent: 'flex-start', flexWrap: 'wrap', display: 'flex', flexDirection: 'row' }}>
                  <InlineFormLabel width={20}>Return Timed Out Field</InlineFormLabel>
                  <Switch
                    css="css"
                    value={timedOutField === null ? true : timedOutField}
                    onChange={this.onTimedOutFieldSwitchChange}
                  />
                </div>
                <div style={{ alignContent: 'flex-start', flexWrap: 'wrap', display: 'flex', flexDirection: 'row' }}>
                  <InlineFormLabel width={20}>Return Total Duration Field</InlineFormLabel>
                  <Switch
                    css="css"
                    value={totalDurationField === null ? true : totalDurationField}
                    onChange={this.onTotalDurationFieldSwitchChange}
                  />
                </div>
                <div style={{ alignContent: 'flex-start', flexWrap: 'wrap', display: 'flex', flexDirection: 'row' }}>
                  <InlineFormLabel width={20}>Return Average Duration Field</InlineFormLabel>
                  <Switch
                    css="css"
                    value={averageDurationField === null ? true : averageDurationField}
                    onChange={this.onAverageDurationFieldSwitchChange}
                  />
                </div>
                <div style={{ alignContent: 'flex-start', flexWrap: 'wrap', display: 'flex', flexDirection: 'row' }}>
                  <InlineFormLabel width={20}>Return Total Time To Recovery Field</InlineFormLabel>
                  <Switch
                    css="css"
                    value={totalTimeToRecoveryField === null ? true : totalTimeToRecoveryField}
                    onChange={this.onTotalTimeToRecoveryFieldSwitchChange}
                  />
                </div>
                <div style={{ alignContent: 'flex-start', flexWrap: 'wrap', display: 'flex', flexDirection: 'row' }}>
                  <InlineFormLabel width={20}>Return Average Time To Recovery Field</InlineFormLabel>
                  <Switch
                    css="css"
                    value={averageTimeToRecoveryField === null ? true : averageTimeToRecoveryField}
                    onChange={this.onAverageTimeToRecoveryFieldSwitchChange}
                  />
                </div>
                <div style={{ alignContent: 'flex-start', flexWrap: 'wrap', display: 'flex', flexDirection: 'column' }}>
                  <div>Enabling the fields below will significantly increase the query time.</div>
                  <div>
                    Note that these values can only be calculated if the release is still available in the Octopus
                    database and has not been cleaned up as part of a retention policy.
                  </div>
                </div>
                <div style={{ alignContent: 'flex-start', flexWrap: 'wrap', display: 'flex', flexDirection: 'row' }}>
                  <InlineFormLabel width={20}>Return Total Cycle Time Field</InlineFormLabel>
                  <Switch
                    css="css"
                    value={totalCycleTimeField === null ? true : totalCycleTimeField}
                    onChange={this.onTotalCycleTimeFieldSwitchChange}
                  />
                </div>
                <div style={{ alignContent: 'flex-start', flexWrap: 'wrap', display: 'flex', flexDirection: 'row' }}>
                  <InlineFormLabel width={20}>Return Average Cycle Time Field</InlineFormLabel>
                  <Switch
                    css="css"
                    value={averageCycleTimeField === null ? true : averageCycleTimeField}
                    onChange={this.onAverageCycleTimeFieldSwitchChange}
                  />
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    );
  }
}
