import React, { ChangeEvent } from 'react';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { InlineField, InlineSwitch, Input, Select } from '@grafana/ui';

import { DataSource } from '../datasource';
import { DEFAULT_QUERY, MyDataSourceOptions, MyQuery } from '../types';
import { EntityFilterSelect } from './EntityFilterSelect';

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

const formatOptions: Array<SelectableValue<string>> = [
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
  { value: 'deployments', label: 'deployments table' },
];

const taskStateOptions: Array<SelectableValue<string>> = [
  { value: 'Success', label: 'Success' },
  { value: 'Failed', label: 'Failed' },
  { value: 'Cancelled', label: 'Cancelled' },
  { value: 'TimedOut', label: 'TimedOut' },
];

interface FieldSwitch {
  label: string;
  key: keyof MyQuery;
}

const timeSeriesSwitches: FieldSwitch[] = [
  { label: 'Return Success Field', key: 'successField' },
  { label: 'Return Failure Field', key: 'failureField' },
  { label: 'Return Cancelled Field', key: 'cancelledField' },
  { label: 'Return Timed Out Field', key: 'timedOutField' },
  { label: 'Return Total Duration Field', key: 'totalDurationField' },
  { label: 'Return Average Duration Field', key: 'averageDurationField' },
  { label: 'Return Total Time To Recovery Field', key: 'totalTimeToRecoveryField' },
  { label: 'Return Average Time To Recovery Field', key: 'averageTimeToRecoveryField' },
];

const leadTimeSwitches: FieldSwitch[] = [
  { label: 'Return Total Deployment Lead Time Field', key: 'totalCycleTimeField' },
  { label: 'Return Average Deployment Lead Time Field', key: 'averageCycleTimeField' },
];

export function QueryEditor({ datasource, query, onChange, onRunQuery }: Props) {
  const merged: MyQuery = { ...DEFAULT_QUERY, ...query };
  const { format, spaceName, projectName, environmentName, channelName, tenantName, releaseVersion, taskState } =
    merged;

  const onFormatChange = (value: SelectableValue<string>) => {
    onChange({ ...merged, format: value.value });
    onRunQuery();
  };

  const onFilterChange = (key: keyof MyQuery) => (value: string) => {
    onChange({ ...merged, [key]: value });
    onRunQuery();
  };

  const onTextChange = (key: keyof MyQuery) => (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...merged, [key]: event.target.value });
  };

  const onSwitchChange = (key: keyof MyQuery) => (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...merged, [key]: event.target.checked });
    onRunQuery();
  };

  const entityFilter = (label: string, key: keyof MyQuery, entityType: string, value: string | undefined) => (
    <EntityFilterSelect
      id={`query-editor-${key}`}
      label={label}
      entityType={entityType}
      datasource={datasource}
      value={value}
      onChange={onFilterChange(key)}
      spaceName={spaceName}
      placeholder="All"
    />
  );

  const switchField = ({ label, key }: FieldSwitch) => (
    <InlineField key={key} label={label} labelWidth={40}>
      <InlineSwitch
        id={`query-editor-${key}`}
        value={merged[key] === undefined ? true : !!merged[key]}
        onChange={onSwitchChange(key)}
      />
    </InlineField>
  );

  return (
    <>
      <InlineField label="Result Format" labelWidth={40} grow>
        <Select
          inputId="query-editor-format"
          value={formatOptions.find((f) => f.value === format) || formatOptions[0]}
          options={formatOptions}
          onChange={onFormatChange}
          width={40}
        />
      </InlineField>
      <EntityFilterSelect
        id="query-editor-spaceName"
        label="Space Name Filter"
        entityType="spaces"
        datasource={datasource}
        value={spaceName}
        onChange={onFilterChange('spaceName')}
        placeholder="Default space"
      />
      {(format === 'timeseries' || format === 'table') && (
        <>
          {entityFilter('Project Name Filter', 'projectName', 'projects', projectName)}
          {entityFilter('Environment Name Filter', 'environmentName', 'environments', environmentName)}
          {entityFilter('Channel Name Filter', 'channelName', 'channels', channelName)}
          {entityFilter('Tenant Name Filter', 'tenantName', 'tenants', tenantName)}
          <InlineField label="Release Version Filter" labelWidth={40} grow>
            <Input
              id="query-editor-releaseVersion"
              value={releaseVersion || ''}
              onChange={onTextChange('releaseVersion')}
              onBlur={onRunQuery}
              width={40}
            />
          </InlineField>
          <InlineField label="Task State Filter" labelWidth={40} grow>
            <Select
              inputId="query-editor-taskState"
              value={
                taskState
                  ? (taskStateOptions.find((o) => o.value === taskState) ?? { label: taskState, value: taskState })
                  : null
              }
              options={taskStateOptions}
              onChange={(selected) => onFilterChange('taskState')(selected?.value ?? '')}
              allowCustomValue
              isClearable
              placeholder="All"
              width={40}
            />
          </InlineField>
          {format === 'timeseries' && (
            <>
              {timeSeriesSwitches.map(switchField)}
              <p>
                Enabling the fields below will significantly increase the query time. Note that these values can only be
                calculated if the release is still available in the Octopus database and has not been cleaned up as part
                of a retention policy.
              </p>
              {leadTimeSwitches.map(switchField)}
            </>
          )}
        </>
      )}
    </>
  );
}
