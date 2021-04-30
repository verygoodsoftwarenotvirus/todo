import axios, { AxiosResponse } from 'axios';
import format from 'string-format';
import { backendRoutes } from '../constants/routes';
import { Logger } from '../logger';
import type {
  Account,
  AccountCreationInput,
  AccountList,
  AuditLogEntry,
  QueryFilter,
} from '../types';

const logger = new Logger().withDebugValue(
  'source',
  'src/apiClient/accounts.ts',
);

export function fetchListOfAccounts(
  qf: QueryFilter,
  adminMode: boolean = false,
): Promise<AxiosResponse<AccountList>> {
  const outboundURLParams = qf.toURLSearchParams();

  if (adminMode) {
    outboundURLParams.set('admin', 'true');
  }

  //

  return axios.get(
    `${backendRoutes.GET_ACCOUNTS}?${outboundURLParams.toString()}`,
  );
}

export function createAccount(
  account: AccountCreationInput,
): Promise<AxiosResponse<Account>> {
  return axios.post(backendRoutes.CREATE_ACCOUNT, account);
}

export function fetchAccount(id: number): Promise<AxiosResponse<Account>> {
  return axios.get(format(backendRoutes.INDIVIDUAL_ACCOUNT, id.toString()));
}

export function saveAccount(account: Account): Promise<AxiosResponse<Account>> {
  return axios.put(
    format(backendRoutes.INDIVIDUAL_ACCOUNT, account.id.toString()),
    account,
  );
}

export function deleteAccount(id: number): Promise<AxiosResponse> {
  return axios.delete(format(backendRoutes.INDIVIDUAL_ACCOUNT, id.toString()));
}

export function fetchAuditLogEntriesForAccount(
  id: number,
): Promise<AxiosResponse<AuditLogEntry[]>> {
  return axios.get(
    format(backendRoutes.INDIVIDUAL_ACCOUNT_AUDIT_LOG, id.toString()),
  );
}
