import React from 'react';
import { SelectableValue } from '@grafana/data';
import { InlineField, Select } from '@grafana/ui';

import { DataSource } from '../datasource';
import { MyVariableQuery } from '../types';
import { EntityFilterSelect } from './EntityFilterSelect';

const entityOptions: Array<SelectableValue<string>> = [
  { value: 'spaces', label: 'spaces' },
  { value: 'environments', label: 'environments' },
  { value: 'tenants', label: 'tenants' },
  { value: 'channels', label: 'channels' },
  { value: 'projects', label: 'projects' },
];

interface VariableQueryProps {
  datasource: DataSource;
  query: MyVariableQuery;
  onChange: (query: MyVariableQuery, definition?: string) => void;
}

export function VariableQueryEditor({ datasource, onChange, query }: VariableQueryProps) {
  const saveQuery = (updated: MyVariableQuery) => {
    onChange(updated, updated.entityName === 'spaces' ? 'spaces' : `${updated.spaceName}: ${updated.entityName}`);
  };

  const onEntityChange = (value: SelectableValue<string>) => {
    saveQuery({ ...query, entityName: value.value || '' });
  };

  const onSpaceNameChange = (value: string) => {
    saveQuery({ ...query, spaceName: value });
  };

  return (
    <>
      <InlineField label="Entity Type" labelWidth={20}>
        <Select
          inputId="variable-query-editor-entity"
          value={entityOptions.find((f) => f.value === query.entityName)}
          options={entityOptions}
          onChange={onEntityChange}
          width={40}
        />
      </InlineField>
      {query.entityName && query.entityName !== 'spaces' && (
        <EntityFilterSelect
          id="variable-query-editor-space"
          label="Space Name"
          entityType="spaces"
          datasource={datasource}
          value={query.spaceName}
          onChange={onSpaceNameChange}
          placeholder="Default space"
          labelWidth={20}
        />
      )}
    </>
  );
}
