import React, { ChangeEvent } from 'react';
import { InlineField, Input } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { ODLDataSourceOptions } from '../types';

interface Props extends DataSourcePluginOptionsEditorProps<ODLDataSourceOptions> {}

export function ConfigEditor(props: Props) {
  const { onOptionsChange, options } = props;
  
  const onBaseUrlChange = (event: ChangeEvent<HTMLInputElement>) => {
    options.jsonData.baseUrl = event.target.value
    onOptionsChange( options );
  };

  console.log(options.jsonData.baseUrl)
  return (
    <div className="gf-form-group">
      <InlineField label="Base URL" labelWidth={12}>
        <Input
          onChange={onBaseUrlChange}
          value={options.jsonData.baseUrl}
          placeholder="http://<host>:<port>"
          width={100}
        />
      </InlineField>
    </div>
  );
}
