import { backendRoutes } from '@/constants/routes';
import { Logger } from '@/logger';
import type { AuditLogEntry, QueryFilter, User } from '@/types';
import axios, { AxiosResponse } from 'axios';
import format from 'string-format';

const logger = new Logger().withDebugValue('source', 'src/apiClient/users.ts');

export function fetchListOfUsers(
  qf: QueryFilter,
  adminMode: boolean = false,
): Promise<AxiosResponse> {
  const outboundURLParams = qf.toURLSearchParams();

  if (adminMode) {
    outboundURLParams.set('admin', 'true');
  }

  const uri = `/api/v1/users?${outboundURLParams.toString()}`;

  return axios.get(uri, { withCredentials: true });
}

export function fetchUser(userID: number): Promise<AxiosResponse<User>> {
  const uri = format(backendRoutes.INDIVIDUAL_USER, userID.toString());
  return axios.get(uri);
}

export function deleteUser(id: number): Promise<AxiosResponse> {
  const uri = format(backendRoutes.INDIVIDUAL_USER, id.toString());
  return axios.delete(uri);
}

export function fetchAuditLogEntriesForUser(
  id: number,
): Promise<AxiosResponse<AuditLogEntry[]>> {
  const uri = format(backendRoutes.INDIVIDUAL_USER_AUDIT_LOG, id.toString());
  return axios.get(uri);
}
