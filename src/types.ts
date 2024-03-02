import { DataSourceJsonData, SelectableValue } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

export interface ColumnType {
  name:             string;
  path:             string;
  regex:            string;
  regexEnabled:     boolean;
  converter:        string;
  converterEnabled: boolean;
}

export interface FilterType {
  field:            string;
  when:             string;
  operation:        string;
  value:            string;
}

export interface ODLQuery extends DataQuery {
  uri:              string;
  loopPath:         string;
  columns:          ColumnType[];
  filters:          FilterType[];
}

export const DEFAULT_QUERY: Partial<ODLQuery> = {
  uri: '/rests/data/network-topology:network-topology',
  loopPath: '$["network-topology:network-topology"]["topology"][0]["link"]',
  columns: 
  [
    {
      name:             'source',
      path:             '$["source"]["source-node"]',
      regex:            'router=(\\d+)',
      regexEnabled:     true,
      converter:        'int2ip',
      converterEnabled: true,
    },
    {
      name:             'target',
      path:             '$["destination"]["dest-node"]',
      regex:            'router=(\\d+)',
      regexEnabled:     true,
      converter:        'int2ip',
      converterEnabled: true,
    },
  ],
  filters:
  [
    {
      field:      'source',
      when:       'raw',
      operation:  '!regexMatch',
      value:      '\\d+:\\d+',
    }
  ]
};

export const conversionOptions: SelectableValue[] = [{label: 'Integer to IP',    value: 'int2ip'}, 
                                                     {label: 'None',             value: 'none'},
                                                     {label: 'SUM',              value: 'sum'}
                                                    ]
export const whenOptions: SelectableValue[] =       [{label: 'Raw value',        value: 'raw'},
                                                     {label: 'After regex',      value: 'regex'},
                                                     {label: 'After conversion', value: 'conversion'}
                                                      ];
export const filterOptions: SelectableValue[] =     [{label: 'Equals',           value: 'equals'},
                                                     {label: 'Greater than',     value: 'gt'},
                                                     {label: 'Less than',        value: 'lt'},
                                                     {label: 'Not equals',       value: '!equals'},
                                                     {label: 'Regex match',      value: 'regexMatch'},
                                                     {label: 'Regex not match',  value: '!regexMatch'},
                                                    ];


/**
 * These are options configured for each DataSource instance
 */
export interface ODLDataSourceOptions extends DataSourceJsonData {
  baseUrl: string;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface ODLSecureJsonData {
  password?: string;
}
