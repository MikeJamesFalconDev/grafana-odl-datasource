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

  // const onUserChange = (event: ChangeEvent<HTMLInputElement>) => {
  //   options.jsonData.user = event.target.value
  //   onOptionsChange( options );
  // };

  // // Secure field (only sent to the backend)
  // const onAPIPasswordChange = (event: ChangeEvent<HTMLInputElement>) => {
  //   onOptionsChange({
  //     ...options,
  //     secureJsonData: {
  //       apiKey: event.target.value,
  //     },
  //   });
  // };

  // const onResetAPIPassword = () => {
  //   onOptionsChange({
  //     ...options,
  //     secureJsonFields: {
  //       ...options.secureJsonFields,
  //       apiKey: false,
  //     },
  //     secureJsonData: {
  //       ...options.secureJsonData,
  //       apiKey: '',
  //     },
  //   });
  // };

//   <InlineField label="User" labelWidth={12}>
//   <Input
//     onChange={onUserChange}
//     value={jsonData.user}
//     placeholder="ODL API user name"
//     width={40}
//   />
// </InlineField>
// <InlineField label="Password" labelWidth={12}>
//   <SecretInput
//     isConfigured={(secureJsonFields && secureJsonFields.apiKey) as boolean}
//     value={secureJsonData.password}
//     placeholder="ODL API password"
//     width={40}
//     onReset={onResetAPIPassword}
//     onChange={onAPIPasswordChange}
//   />
// </InlineField>
  // const { jsonData, secureJsonFields } = options;
  // const secureJsonData = (options.secureJsonData || {}) as ODLSecureJsonData;
// || '192.168.230.136'
  console.log(options.jsonData.baseUrl)
  return (
    <div className="gf-form-group">
      <InlineField label="Baseurl" labelWidth={12}>
        <Input
          onChange={onBaseUrlChange}
          value={options.jsonData.baseUrl}
          placeholder="http://192.168.230.136:8181"
          width={100}
        />
      </InlineField>
    </div>
  );
}
