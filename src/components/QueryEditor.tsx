import React, { ChangeEvent } from 'react';
import { Button, Card, InlineField, Input, CollapsableSection, Select, ActionMeta } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataSource } from '../datasource';
import { ODLDataSourceOptions, ODLQuery } from '../types';

type Props = QueryEditorProps<DataSource, ODLQuery, ODLDataSourceOptions>;

export function QueryEditor({ query, onChange, onRunQuery }: Props) {

  const onUriChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, uri: event.target.value });
    onRunQuery();
  };

  const onLoopPathChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, loopPath: event.target.value });
    // executes the query
    onRunQuery();
  };

  const onColumnNameChange = (index: number, event: ChangeEvent<HTMLInputElement>) => {
    console.log(query.columns[index].name)
    query.columns[index].name = event.target.value
    onChange(query);
   // executes the query
   onRunQuery();
  };

  const onColumnPathChange = (index: number, event: ChangeEvent<HTMLInputElement>) => {
    console.log(query.columns[index].path)
    query.columns[index].path = event.target.value
    onChange(query);
   // executes the query
   onRunQuery();
  };

  const onColumnRegexChange = (index: number, event: ChangeEvent<HTMLInputElement>) => {
    console.log(query.columns[index].path)
    query.columns[index].regex = event.target.value
    onChange(query);
   // executes the query
   onRunQuery();
  };

  const onConverterChange = (index: number, value: SelectableValue<string> ,action: ActionMeta) => {
    console.log(query.columns[index].path)
    if (value.value) {
      query.columns[index].converter = value.value
      onChange(query);
      onRunQuery();
    }
  };

  const addColumn = () => {
    columns.push({
      name:       '',
      path:       '',
      regex:      '',
      converter:  'none',
    })
    onChange(query)
  }

  const deleteColumn = () => {
    columns.pop()
    onChange(query)
  }

  const { uri, loopPath, columns } = query;
  const conversionOptions: SelectableValue[] = [{label: 'Integer to IP', value: 'int2ip'}, {label: 'None', value: 'none'}]

  return (
    <div>
      <Card>
        <InlineField label="URI" labelWidth={16} tooltip="ODL REST call relative path">
          <Input onChange={onUriChange} value={uri} width={150}/>
        </InlineField>
      </Card>
      <Card>
        <InlineField label="Loop Path" labelWidth={16} tooltip="Path within the API response that renders each row.">
          <Input onChange={onLoopPathChange} width={150} value={loopPath || 'network-topology:network-topology/topology[0]/link'} />
        </InlineField>
      </Card>
      <CollapsableSection label='Columns' isOpen={true}>
        {columns.map((column, index, columns) => 
            <Card key={index} className='x-flex'>
              <InlineField labelWidth={18} label='Name' tooltip={'The name for the column where the value specified by the path will be put'}>
                <Input onChange={(e: ChangeEvent<HTMLInputElement>) => onColumnNameChange(index,e)} width={100} value={column.name}/>
              </InlineField>
              <InlineField labelWidth={18} label='Path' tooltip={'Path within the JSON respons to find the value for this column'}>
                <Input onChange={(e: ChangeEvent<HTMLInputElement>) => onColumnPathChange(index, e)} width={100} value={column.path}/>
              </InlineField>
              <InlineField labelWidth={18} label='Regex' tooltip={'Regular expression to apply to this column to extract the value. If empty the selected value is sent as is'}>
                <Input onChange={(e: ChangeEvent<HTMLInputElement>) => onColumnRegexChange(index, e)} width={100} value={column.regex}/>
              </InlineField>
              <InlineField labelWidth={18} label='Conversion' tooltip={'Conversion function to apply to the value. If a regex is supplied, the regex is applied first then the conviersion function.'}>
                <Select onChange={(value: SelectableValue<string> ,action: ActionMeta) => onConverterChange(index,value, action)} 
                      width={100} 
                      value={conversionOptions.find(value => value.value === column.converter)}
                      closeMenuOnSelect={true}
                      options={conversionOptions}
                />
              </InlineField>
            </Card>
        )}
      </CollapsableSection>
      <Button onClick={addColumn}>+ Column</Button>
      <Button onClick={deleteColumn}>- Column</Button>
    </div>
  );
}
