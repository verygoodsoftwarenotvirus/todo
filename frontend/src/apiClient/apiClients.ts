import { backendRoutes } from '../constants/routes';
import { Logger } from '../logger';
import type {
  APIClient,
  APIClientCreationInput,
  APIClientList,
  AuditLogEntry,
  QueryFilter,
} from '../types';
import axios, { AxiosResponse } from 'axios';
import format from 'string-format';

const logger = new Logger().withDebugValue(
  'source',
  'src/apiClient/accounts.ts',
);

export function fetchListOfAPIClients(
  qf: QueryFilter,
  adminMode: boolean = false,
): Promise<AxiosResponse<APIClientList>> {
  const outboundURLParams = qf.toURLSearchParams();

  if (adminMode) {
    outboundURLParams.set('admin', 'true');
  }

  return axios.get(
    `${backendRoutes.GET_ACCOUNTS}?${outboundURLParams.toString()}`,
  );
}

export function createAPIClient(
  account: APIClientCreationInput,
): Promise<AxiosResponse<APIClient>> {
  return axios.post(backendRoutes.CREATE_ACCOUNT, account);
}

export function fetchAPIClient(id: number): Promise<AxiosResponse<APIClient>> {
  return axios.get(format(backendRoutes.INDIVIDUAL_ACCOUNT, id.toString()));
}

export function saveAPIClient(
  account: APIClient,
): Promise<AxiosResponse<APIClient>> {
  return axios.put(
    format(backendRoutes.INDIVIDUAL_ACCOUNT, account.id.toString()),
    account,
  );
}

export function deleteAPIClient(id: number): Promise<AxiosResponse> {
  return axios.delete(format(backendRoutes.INDIVIDUAL_ACCOUNT, id.toString()));
}

export function fetchAuditLogEntriesForAPIClient(
  id: number,
): Promise<AxiosResponse<AuditLogEntry[]>> {
  return axios.get(
    format(backendRoutes.INDIVIDUAL_ACCOUNT_AUDIT_LOG, id.toString()),
  );
}
