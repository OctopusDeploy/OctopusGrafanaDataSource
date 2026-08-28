import React from 'react';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { InlineField, Select } from '@grafana/ui';

import { DataSource } from '../datasource';
import {
  ANNOTATION_FORMAT_DEPLOYMENTS,
  ANNOTATION_FORMAT_DEPLOYMENT_REPORT,
  MyDataSourceOptions,
  MyQuery,
} from '../types';
import { EntityFilterSelect } from './EntityFilterSelect';

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

const annotationFormatOptions: Array<SelectableValue<string>> = [
  {
    value: ANNOTATION_FORMAT_DEPLOYMENT_REPORT,
    label: 'Deployment reports',
    description: 'The start and end time of completed deployments',
  },
  {
    value: ANNOTATION_FORMAT_DEPLOYMENTS,
    label: 'Deployments',
    description: 'The start time of recent deployments, including those in progress',
  },
];

export function AnnotationQueryEditor({ datasource, query, onChange }: Props) {
  const format = query.format || ANNOTATION_FORMAT_DEPLOYMENT_REPORT;

  const onFormatChange = (value: SelectableValue<string>) => {
    onChange({ ...query, format: value.value });
  };

  const onFilterChange = (key: keyof MyQuery) => (value: string) => {
    onChange({ ...query, [key]: value });
  };

  return (
    <>
      <InlineField label="Annotation Type" labelWidth={20}>
        <Select
          inputId="annotation-editor-format"
          value={annotationFormatOptions.find((f) => f.value === format) || annotationFormatOptions[0]}
          options={annotationFormatOptions}
          onChange={onFormatChange}
          width={40}
        />
      </InlineField>
      <EntityFilterSelect
        id="annotation-editor-space"
        label="Space Name"
        entityType="spaces"
        datasource={datasource}
        value={query.spaceName}
        onChange={onFilterChange('spaceName')}
        placeholder="Default space"
        labelWidth={20}
      />
      <EntityFilterSelect
        id="annotation-editor-environment"
        label="Environment Name"
        entityType="environments"
        datasource={datasource}
        value={query.environmentName}
        onChange={onFilterChange('environmentName')}
        spaceName={query.spaceName}
        placeholder="All"
        labelWidth={20}
      />
      <EntityFilterSelect
        id="annotation-editor-project"
        label="Project Name"
        entityType="projects"
        datasource={datasource}
        value={query.projectName}
        onChange={onFilterChange('projectName')}
        spaceName={query.spaceName}
        placeholder="All"
        labelWidth={20}
      />
    </>
  );
}
