import { backendRoutes } from '@/constants/routes';
import { Logger } from '@/logger';
import type {
  AuditLogEntry,
  OAuth2Client,
  OAuth2ClientCreationInput,
  OAuth2ClientList,
  QueryFilter,
} from '@/types';
import axios, { AxiosResponse } from 'axios';
import format from 'string-format';

const logger = new Logger().withDebugValue(
  'source',
  'src/apiClient/oauth2clients.ts',
);

export function fetchListOfOAuth2Clients(
  qf: QueryFilter,
  adminMode: boolean = false,
): Promise<AxiosResponse<OAuth2ClientList>> {
  const outboundURLParams = qf.toURLSearchParams();

  if (adminMode) {
    outboundURLParams.set('admin', 'true');
  }

  const uri = `${
    backendRoutes.GET_OAUTH2_CLIENTS
  }?${outboundURLParams.toString()}`;
  return axios.get(uri);
}

export function createOAuth2Client(
  oauth2client: OAuth2ClientCreationInput,
): Promise<AxiosResponse<OAuth2Client>> {
  return axios.post(backendRoutes.CREATE_OAUTH2_CLIENT, oauth2client);
}

export function fetchOAuth2Client(
  id: number,
): Promise<AxiosResponse<OAuth2Client>> {
  const uri = format(backendRoutes.INDIVIDUAL_OAUTH2_CLIENT, id.toString());
  return axios.get(uri);
}

export function deleteOAuth2Client(id: number): Promise<AxiosResponse> {
  const uri = format(backendRoutes.INDIVIDUAL_OAUTH2_CLIENT, id.toString());
  return axios.delete(uri);
}

export function fetchAuditLogEntriesForOAuth2Client(
  id: number,
): Promise<AxiosResponse<AuditLogEntry[]>> {
  const uri = format(
    backendRoutes.INDIVIDUAL_OAUTH2_CLIENT_AUDIT_LOG,
    id.toString(),
  );
  return axios.get(uri);
}
