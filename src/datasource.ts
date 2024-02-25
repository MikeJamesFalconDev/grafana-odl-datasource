import { DataSourceInstanceSettings, CoreApp } from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';

import { ODLQuery, ODLDataSourceOptions, DEFAULT_QUERY } from './types';

export class DataSource extends DataSourceWithBackend<ODLQuery, ODLDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<ODLDataSourceOptions>) {
    super(instanceSettings);
  }

  getDefaultQuery(_: CoreApp): Partial<ODLQuery> {
    return DEFAULT_QUERY;
  }
}
