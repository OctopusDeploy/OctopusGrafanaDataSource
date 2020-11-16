import React, { useState } from 'react';
import {SelectableValue} from "@grafana/data";
import {LegacyForms} from "@grafana/ui";
const {Select} = LegacyForms;

export interface MyVariableQuery {
  spaceName: string;
  entityName: string;
}

interface VariableQueryProps {
  query: MyVariableQuery;
  onChange: (query: MyVariableQuery, definition: string) => void;
}

export const VariableQueryEditor: React.FC<VariableQueryProps> = ({ onChange, query }) => {
  const [state, setState] = useState(query);

  const formatOptions = [
    {value: "spaces", label: "spaces"},
    {value: "environments", label: "environments"},
    {value: "tenants", label: "tenants"},
    {value: "channels", label: "channels"},
    {value: "projects", label: "projects"}];

  const saveQuery = () => {
    onChange(state, `${state.spaceName}: ${state.entityName}`);
  };

  const handleChange = (event: React.FormEvent<HTMLInputElement>) =>
    setState({
      ...state,
      spaceName: event.currentTarget.value,
    });

  const onEntityTextChange = (value: SelectableValue<string>) => {
    setState({
      ...state,
      entityName: value.value || "",
    });
  };

  return (
    <>
      <div className="gf-form">
        <span className="gf-form-label width-10">Entity Type</span>
        <Select
          value={formatOptions.find(f => f.value == state.entityName)}
          options={formatOptions}
          onChange={onEntityTextChange}
          onBlur={saveQuery}
        />
      </div>
      {state.entityName != "spaces" && state.entityName != "" &&
        <div className="gf-form">
          <span className="gf-form-label width-10">Space Name</span>
          <input
            name="spaceName"
            className="gf-form-input"
            onBlur={saveQuery}
            onChange={handleChange}
            value={state.spaceName}
          />
        </div>
      }
    </>
  );
};
