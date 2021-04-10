import { backendRoutes } from '../constants/routes';
import { Logger } from '../logger';
import type { AuditLogEntry, QueryFilter, User } from '../types';
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

  return axios.get(`/api/v1/users?${outboundURLParams.toString()}`, {
    withCredentials: true,
  });
}

export function fetchUser(userID: number): Promise<AxiosResponse<User>> {
  return axios.get(format(backendRoutes.INDIVIDUAL_USER, userID.toString()));
}

export function deleteUser(id: number): Promise<AxiosResponse> {
  return axios.delete(format(backendRoutes.INDIVIDUAL_USER, id.toString()));
}

export function fetchAuditLogEntriesForUser(
  id: number,
): Promise<AxiosResponse<AuditLogEntry[]>> {
  return axios.get(
    format(backendRoutes.INDIVIDUAL_USER_AUDIT_LOG, id.toString()),
  );
}
