import defaults from 'lodash/defaults';

import React, {ChangeEvent, PureComponent} from 'react';
import {InlineFormLabel, LegacyForms, Switch} from '@grafana/ui';
import {QueryEditorProps, SelectableValue} from '@grafana/data';
import {DataSource} from './DataSource';
import {defaultQuery, MyDataSourceOptions, MyQuery} from './types';

const {FormField, Select} = LegacyForms;

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export class QueryEditor extends PureComponent<Props> {
  onFormatTextChange = (value: SelectableValue<string>) => {
    const {onChange, query} = this.props;
    onChange({...query, format: value.value});
  };

  onSpaceNameTextChange = (event: ChangeEvent<HTMLInputElement>) => {
    const {onChange, query} = this.props;
    onChange({...query, spaceName: event.target.value});
  };

  onProjectNameTextChange = (event: ChangeEvent<HTMLInputElement>) => {
    const {onChange, query} = this.props;
    onChange({...query, projectName: event.target.value});
  };

  onEnvironmentNameTextChange = (event: ChangeEvent<HTMLInputElement>) => {
    const {onChange, query} = this.props;
    onChange({...query, environmentName: event.target.value});
  };

  onTenantNameTextChange = (event: ChangeEvent<HTMLInputElement>) => {
    const {onChange, query} = this.props;
    onChange({...query, tenantName: event.target.value});
  };

  onChannelNameTextChange = (event: ChangeEvent<HTMLInputElement>) => {
    const {onChange, query} = this.props;
    onChange({...query, channelName: event.target.value});
  };

  onReleaseVersionTextChange = (event: ChangeEvent<HTMLInputElement>) => {
    const {onChange, query} = this.props;
    onChange({...query, releaseVersion: event.target.value});
  };

  onTaskSTateTextChange = (event: ChangeEvent<HTMLInputElement>) => {
    const {onChange, query} = this.props;
    onChange({...query, taskState: event.target.value});
  };

  onSuccessFieldSwitchChange = (event: ChangeEvent<HTMLInputElement>) => {
    const {onChange, query} = this.props;
    onChange({...query, successField: event.target.checked});
  };

  onFailureFieldSwitchChange = (event: ChangeEvent<HTMLInputElement>) => {
    const {onChange, query} = this.props;
    onChange({...query, failureField: event.target.checked});
  };

  onCancelledFieldSwitchChange = (event: ChangeEvent<HTMLInputElement>) => {
    const {onChange, query} = this.props;
    onChange({...query, cancelledField: event.target.checked});
  };

  onTimedOutFieldSwitchChange = (event: ChangeEvent<HTMLInputElement>) => {
    const {onChange, query} = this.props;
    onChange({...query, timedOutField: event.target.checked});
  };

  onTotalDurationFieldSwitchChange = (event: ChangeEvent<HTMLInputElement>) => {
    const {onChange, query} = this.props;
    onChange({...query, totalDurationField: event.target.checked});
  };

  onAverageDurationFieldSwitchChange = (event: ChangeEvent<HTMLInputElement>) => {
    const {onChange, query} = this.props;
    onChange({...query, averageDurationField: event.target.checked});
  };

  onTotalTimeToRecoveryFieldSwitchChange = (event: ChangeEvent<HTMLInputElement>) => {
    const {onChange, query} = this.props;
    onChange({...query, totalTimeToRecoveryField: event.target.checked});
  };

  onAverageTimeToRecoveryFieldSwitchChange = (event: ChangeEvent<HTMLInputElement>) => {
    const {onChange, query} = this.props;
    onChange({...query, averageTimeToRecoveryField: event.target.checked});
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
      averageTimeToRecoveryField
    } = query;
    const formatOptions = [
      {value: "timeseries", label: "deployments time series"},
      {value: "table", label: "deployments table"},
      {value: "environments", label: "environments table"},
      {value: "tenants", label: "tenants table"},
      {value: "channels", label: "channels table"},
      {value: "projects", label: "environments table"}];

    return (
      <div className="gf-form" style={{flexDirection: "column"}}>
        <div style={{alignContent: "flex-start", flexWrap: "wrap", display: "flex", flexDirection: "row"}}>
          <InlineFormLabel width={20}>Result Format</InlineFormLabel>
          <Select
            value={formatOptions.find(f => f.value == format) || formatOptions.find(f => f.value == "timeseries")}
            options={formatOptions}
            onChange={this.onFormatTextChange}
          />
        </div>
        <FormField
          labelWidth={8}
          value={spaceName || ''}
          onChange={this.onSpaceNameTextChange}
          label="Space Name"
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
        <FormField
          labelWidth={8}
          value={taskState || ''}
          onChange={this.onTaskSTateTextChange}
          label="Task State"
        />
        <div style={{alignContent: "flex-start", flexWrap: "wrap", display: "flex", flexDirection: "row"}}>
          <InlineFormLabel width={20}>Success field</InlineFormLabel>
          <Switch
            css="css"
            value={successField == null ? true : successField}
            onChange={this.onSuccessFieldSwitchChange}
          />
        </div>
        <div style={{alignContent: "flex-start", flexWrap: "wrap", display: "flex", flexDirection: "row"}}>
          <InlineFormLabel width={20}>Failure field</InlineFormLabel>
          <Switch
            css="css"
            value={failureField == null ? true : failureField}
            onChange={this.onFailureFieldSwitchChange}
          />
        </div>
        <div style={{alignContent: "flex-start", flexWrap: "wrap", display: "flex", flexDirection: "row"}}>
          <InlineFormLabel width={20}>Cancelled field</InlineFormLabel>
          <Switch
            css="css"
            value={cancelledField == null ? true : cancelledField}
            onChange={this.onCancelledFieldSwitchChange}
          />
        </div>
        <div style={{alignContent: "flex-start", flexWrap: "wrap", display: "flex", flexDirection: "row"}}>
          <InlineFormLabel width={20}>Timed Out field</InlineFormLabel>
          <Switch
            css="css"
            value={timedOutField == null ? true : timedOutField}
            onChange={this.onTimedOutFieldSwitchChange}
          />
        </div>
        <div style={{alignContent: "flex-start", flexWrap: "wrap", display: "flex", flexDirection: "row"}}>
          <InlineFormLabel width={20}>Total Duration Field</InlineFormLabel>
          <Switch
            css="css"
            value={totalDurationField == null ? true : totalDurationField}
            onChange={this.onTotalDurationFieldSwitchChange}
          />
        </div>
        <div style={{alignContent: "flex-start", flexWrap: "wrap", display: "flex", flexDirection: "row"}}>
          <InlineFormLabel width={20}>Average Duration Field</InlineFormLabel>
          <Switch
            css="css"
            value={averageDurationField == null ? true : averageDurationField}
            onChange={this.onAverageDurationFieldSwitchChange}
          />
        </div>
        <div style={{alignContent: "flex-start", flexWrap: "wrap", display: "flex", flexDirection: "row"}}>
          <InlineFormLabel width={20}>Total Time To Recovery Field</InlineFormLabel>
          <Switch
            css="css"
            value={totalTimeToRecoveryField == null ? true : totalTimeToRecoveryField}
            onChange={this.onTotalTimeToRecoveryFieldSwitchChange}
          />
        </div>
        <div style={{alignContent: "flex-start", flexWrap: "wrap", display: "flex", flexDirection: "row"}}>
          <InlineFormLabel width={20}>Average Time To Recovery Field</InlineFormLabel>
          <Switch
            css="css"
            value={averageTimeToRecoveryField == null ? true : averageTimeToRecoveryField}
            onChange={this.onAverageTimeToRecoveryFieldSwitchChange}
          />
        </div>
      </div>
    );
  }
}
