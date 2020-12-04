import { backendRoutes } from '@/constants/routes';
import { Logger } from '@/logger';
import type { AuditLogEntry, AuditLogEntryList, QueryFilter } from '@/types';
import axios, { AxiosResponse } from 'axios';
import format from 'string-format';

const logger = new Logger().withDebugValue(
  'source',
  'src/apiClient/audit_log_entries.ts',
);

export function fetchListOfAuditLogEntries(
  qf: QueryFilter,
  adminMode: boolean = false,
): Promise<AxiosResponse<AuditLogEntryList>> {
  const outboundURLParams = qf.toURLSearchParams();

  if (adminMode) {
    outboundURLParams.set('admin', 'true');
  }

  const uri = `${
    backendRoutes.GET_AUDIT_LOG_ENTRIES
  }?${outboundURLParams.toString()}`;
  return axios.get(uri);
}

export function fetchAuditLogEntry(
  id: number,
): Promise<AxiosResponse<AuditLogEntry>> {
  const uri = format(backendRoutes.INDIVIDUAL_AUDIT_LOG_ENTRY, id.toString());
  return axios.get(uri);
}
