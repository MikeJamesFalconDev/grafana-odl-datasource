import React, { ChangeEvent } from 'react';
import { Button, Card, InlineField, Input, CollapsableSection, Select, ActionMeta, useStyles2, Checkbox, InlineFieldRow } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataSource } from '../datasource';
import { ColumnType, ODLDataSourceOptions, ODLQuery, conversionOptions, filterOptions, whenOptions } from '../types';
import { css, cx } from '@emotion/css';


// TODO add checkbox to disable regex and conversion. If unchecked, the values are set empty. Somehow avoid loosing the values.




type Props = QueryEditorProps<DataSource, ODLQuery, ODLDataSourceOptions>;

const getStyles = () => {
  return {
    wrapper: css`
      font-family: Open Sans;
      position: relative;
    `,
    svg: css`
      position: absolute;
      top: 0;
      left: 0;
    `,
    textBox: css`
      position: absolute;
      bottom: 0;
      left: 0;
      padding: 10px;
    `,
  };
};


export function QueryEditor({ query, onChange, onRunQuery }: Props) {
  const styles = useStyles2(getStyles);

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
    query.columns[index].name = event.target.value
    onChange(query);
   // executes the query
   onRunQuery();
  };

  const onColumnPathChange = (index: number, event: ChangeEvent<HTMLInputElement>) => {
    query.columns[index].path = event.target.value
    onChange(query);
   // executes the query
   onRunQuery();
  };

  const onColumnRegexChange = (index: number, event: ChangeEvent<HTMLInputElement>) => {
    columns[index].regex = event.target.value
    onChange(query);
   // executes the query
   onRunQuery();
  };

  const onConverterChange = (index: number, value: SelectableValue<string> ,action: ActionMeta) => {
    if (value.value) {
      query.columns[index].converter = value.value
      onChange(query);
      onRunQuery();
    }
  };

  const onRegexEnabledChange = (index: number) => {
    query.columns[index].regexEnabled = !query.columns[index].regexEnabled
    onChange(query);
   // executes the query
   onRunQuery();
  };

  const onConverterEnabledChange = (index: number) => {
    console.log(query.columns[index].path)
    query.columns[index].converterEnabled = !query.columns[index].converterEnabled
    onChange(query);
   // executes the query
   onRunQuery();
  };

  const addColumn = () => {
    columns.push({
      name:             '',
      path:             '',
      regex:            '',
      regexEnabled:     true,
      converter:        'none',
      converterEnabled: true,
    })
    onChange(query)
  }

  const deleteColumn = () => {
    columns.pop()
    onChange(query)
  }

  const onFilterFieldChange = (index: number, value: SelectableValue<string> ,action: ActionMeta) => {
    if (value.value) {
      query.filters[index].field = value.value
      onChange(query);
      onRunQuery();
    }
  }

  const onFilterWhenChange = (index: number, value: SelectableValue<string> ,action: ActionMeta) => {
    if (value.value) {
      query.filters[index].when = value.value
      onChange(query);
      onRunQuery();
    }
  }

  const onFilterOperationChange = (index: number, value: SelectableValue<string> ,action: ActionMeta) => {
    if (value.value) {
      query.filters[index].operation = value.value
      onChange(query);
      onRunQuery();
    }
  }

  const onFilterValueChange = (index: number, event: ChangeEvent<HTMLInputElement>) => {
    query.filters[index].value = event.target.value
    onChange(query);
   // executes the query
   onRunQuery();
  };

  const addFilter = () => {
    filters.push({
      field:      '',
      when:       '',
      operation:  '',
      value:      '',
    })
    onChange(query)
  }

  const deleteFilter = () => {
    filters.pop()
    onChange(query)
  }

  const toSelectValues = (columns: ColumnType[]): SelectableValue[] => {
    return columns.map((column: ColumnType): SelectableValue<string> => { return {label: column.name, value: column.name }})
  }

  if (!query.filters ) {
    query.filters = [{ field: '', when:'', operation: '', value: '' }]
  }
  const { uri, loopPath, columns, filters } = query;
  

  return (
    <div
    className={cx(
      styles.wrapper,
      css`
        flex: 1
      `
    )}
  >
      <Card>
        <InlineField label="URI" labelWidth={18} tooltip="ODL REST call relative path">
          <Input onChange={onUriChange} value={uri} width={100}/>
        </InlineField>
      </Card>
      <Card>
        <InlineField label="Loop Path" labelWidth={18} tooltip="Path within the API response that renders each row.">
          <Input onChange={onLoopPathChange} width={100} value={loopPath || 'network-topology:network-topology/topology[0]/link'} />
        </InlineField>
      </Card>
      <CollapsableSection label='Columns' isOpen={true}>
        {columns.map((column, index) => 
            <Card key={index} >
              <div>
              <InlineField labelWidth={18} label='Name' tooltip={'The name for the column where the value specified by the path will be put'}>
                <Input onChange={(e: ChangeEvent<HTMLInputElement>) => onColumnNameChange(index,e)} width={100} value={column.name}/>
              </InlineField>
              <InlineField labelWidth={18} label='Path' tooltip={'Path within the JSON respons to find the value for this column'}>
                <Input onChange={(e: ChangeEvent<HTMLInputElement>) => onColumnPathChange(index, e)} width={100} value={column.path}/>
              </InlineField>
              <InlineFieldRow>
                <Checkbox onChange={e => onRegexEnabledChange(index)} value={column.regexEnabled}/>
                <InlineField disabled={!column.regexEnabled} labelWidth={16} label='Regex' tooltip={'Regular expression to apply to this column to extract the value. If empty the selected value is sent as is'}>
                  <Input onChange={(e: ChangeEvent<HTMLInputElement>) => onColumnRegexChange(index, e)} width={100} value={column.regex}/>
                </InlineField>
              </InlineFieldRow>
              <InlineFieldRow>
                <Checkbox onChange={e => onConverterEnabledChange(index)} value={column.converterEnabled}/>
                <InlineField disabled={!column.converterEnabled} labelWidth={16} label='Conversion' tooltip={'Conversion function to apply to the value. If a regex is supplied, the regex is applied first then the conviersion function.'}>
                  <Select onChange={(value: SelectableValue<string>, action: ActionMeta) => onConverterChange(index,value, action)} 
                        width={100} 
                        value={conversionOptions.find(value => value.value === column.converter)}
                        closeMenuOnSelect={true}
                        options={conversionOptions}
                  />
                </InlineField>
              </InlineFieldRow>
              </div>
            </Card>
        )}
        <Button onClick={addColumn}>+ Column</Button>
        <Button onClick={deleteColumn}>- Column</Button>
      </CollapsableSection>
      <CollapsableSection label='Filters' isOpen={false}>
          {
          filters.map((filter,index) =>
            <Card key={index}>
              <div>
              <InlineField labelWidth={18} label='Field' tooltip={'Field used in filter operation'}>
                <Select onChange={(value: SelectableValue<string>, action: ActionMeta) => onFilterFieldChange(index,value, action)} 
                      width={100} 
                      value={toSelectValues(columns).find(value => value.value === filter.field)}
                      closeMenuOnSelect={true}
                      options={toSelectValues(columns)}
                />
              </InlineField>
              <InlineField labelWidth={18} label='When' tooltip={'When to perform matching. Options are to filter on the raw value, after Regex is perormed or after conversion is performed.'}>
                <Select onChange={(value: SelectableValue<string>, action: ActionMeta) => onFilterWhenChange(index,value, action)} 
                      width={100} 
                      value={whenOptions.find(value => value.value === filter.when)}
                      closeMenuOnSelect={true}
                      options={whenOptions}
                />
              </InlineField>
              <InlineField labelWidth={18} label='Filter' tooltip={'Filter criteria for the field'}>
                <Select onChange={(value: SelectableValue<string>, action: ActionMeta) => onFilterOperationChange(index,value, action)} 
                      width={100} 
                      value={filterOptions.find(value => value.value === filter.operation)}
                      closeMenuOnSelect={true}
                      options={filterOptions}
                />
              </InlineField>
              <InlineField labelWidth={18} label='Value' tooltip={'The name for the column where the value specified by the path will be put'}>
                <Input onChange={(e: ChangeEvent<HTMLInputElement>) => onFilterValueChange(index,e)} width={100} value={filter.value}/>
              </InlineField>
              </div>
            </Card>
          )}
        <Button onClick={addFilter}>+ Filter</Button>
        <Button onClick={deleteFilter}>- Filter</Button>
      </CollapsableSection>
    </div>
  );
}
