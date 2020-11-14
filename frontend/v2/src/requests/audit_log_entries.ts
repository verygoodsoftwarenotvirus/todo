import axios, { AxiosResponse } from 'axios';
import format from 'string-format';

import type { QueryFilter } from '@/types';
import { Logger } from '@/logger';
import { backendRoutes } from '@/constants/routes';
import {
  defaultAPIRequestConfig,
  requestLogFunction,
} from '@/requests/defaults';

const logger = new Logger().withDebugValue(
  'source',
  'src/requests/audit_log_entries.ts',
);

export function fetchListOfAuditLogEntries(
  qf: QueryFilter,
  adminMode: boolean = false,
): Promise<AxiosResponse> {
  const outboundURLParams = qf.toURLSearchParams();

  if (adminMode) {
    outboundURLParams.set('admin', 'true');
  }

  const uri = `${
    backendRoutes.GET_AUDIT_LOG_ENTRIES
  }?${outboundURLParams.toString()}`;
  return axios
    .get(uri, defaultAPIRequestConfig)
    .then(requestLogFunction(logger, uri));
}

export function fetchAuditLogEntry(id: number): Promise<AxiosResponse> {
  const uri = format(backendRoutes.INDIVIDUAL_AUDIT_LOG_ENTRY, id.toString());
  return axios
    .get(uri, defaultAPIRequestConfig)
    .then(requestLogFunction(logger, uri));
}
