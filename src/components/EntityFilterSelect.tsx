import React, { useEffect, useState } from 'react';
import { SelectableValue } from '@grafana/data';
import { InlineField, Select } from '@grafana/ui';

import { DataSource } from '../datasource';

/**
 * Loads the names of an entity type from the datasource as select options.
 * Lookup failures (e.g. an unconfigured datasource) resolve to an empty list,
 * leaving the select usable as a free-text field.
 */
export function useEntityOptions(
  datasource: DataSource,
  entityType: string,
  spaceName?: string,
  enabled = true
): Array<SelectableValue<string>> {
  const [options, setOptions] = useState<Array<SelectableValue<string>>>([]);

  useEffect(() => {
    if (!enabled) {
      return;
    }

    let active = true;
    datasource
      .getEntityNames(entityType, spaceName)
      .then((names) => {
        if (active) {
          setOptions(names.map((value) => ({ label: value, value })));
        }
      })
      .catch(() => {
        if (active) {
          setOptions([]);
        }
      });
    return () => {
      active = false;
    };
  }, [datasource, entityType, spaceName, enabled]);

  return options;
}

interface EntityFilterSelectProps {
  id: string;
  label: string;
  entityType: string;
  datasource: DataSource;
  value?: string;
  onChange: (value: string) => void;
  spaceName?: string;
  placeholder?: string;
  labelWidth?: number;
}

/**
 * A select listing the entities of one type, scoped to a space. Custom values
 * are allowed so template variables (e.g. $space) and names from other
 * credentials keep working, and clearing the select returns to the unfiltered
 * default.
 */
export function EntityFilterSelect({
  id,
  label,
  entityType,
  datasource,
  value,
  onChange,
  spaceName,
  placeholder,
  labelWidth,
}: EntityFilterSelectProps) {
  const options = useEntityOptions(datasource, entityType, spaceName);
  const current = value ? (options.find((o) => o.value === value) ?? { label: value, value }) : null;

  return (
    <InlineField label={label} labelWidth={labelWidth ?? 40} grow>
      <Select
        inputId={id}
        value={current}
        options={options}
        onChange={(selected) => onChange(selected?.value ?? '')}
        allowCustomValue
        isClearable
        placeholder={placeholder}
        width={40}
      />
    </InlineField>
  );
}
