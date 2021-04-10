import { backendRoutes } from '../constants/routes';
import { Logger } from '../logger';
import type { AuditLogEntry, AuditLogEntryList, QueryFilter } from '../types';
import axios, { AxiosResponse } from 'axios';
import format from 'string-format';

const logger = new Logger().withDebugValue(
  'source',
  'src/apiClient/auditLogEntries.ts',
);

export function fetchListOfAuditLogEntries(
  qf: QueryFilter,
  adminMode: boolean = false,
): Promise<AxiosResponse<AuditLogEntryList>> {
  const outboundURLParams = qf.toURLSearchParams();

  if (adminMode) {
    outboundURLParams.set('admin', 'true');
  }

  return axios.get(
    `${backendRoutes.GET_AUDIT_LOG_ENTRIES}?${outboundURLParams.toString()}`,
  );
}

export function fetchAuditLogEntry(
  id: number,
): Promise<AxiosResponse<AuditLogEntry>> {
  return axios.get(
    format(backendRoutes.INDIVIDUAL_AUDIT_LOG_ENTRY, id.toString()),
  );
}
