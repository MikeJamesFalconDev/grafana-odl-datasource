import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

export interface ColumnType {
    name:       string;
    path:       string;
    regex:      string;
    converter:  string;
}

export interface ODLQuery extends DataQuery {
  uri: string;
  loopPath: string;
  columns: ColumnType[];
}

export const DEFAULT_QUERY: Partial<ODLQuery> = {
  uri: '/rests/data/network-topology:network-topology',
  loopPath: 'network-topology:network-topology/topology[0]/link',
  columns: 
  [
    {
      name:       'source',
      path:       'source/source-node',
      regex:      '.*router=(\d+).*/$1',
      converter:  'int2ip',
    },
    {
      name:       'target',
      path:       'destination/dest-node',
      regex:      '.*router=(\d+).*/$1',
      converter:  'int2ip',
    },
  ],
};

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
