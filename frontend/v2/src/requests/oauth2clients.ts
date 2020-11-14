import axios, { AxiosResponse } from 'axios';
import format from 'string-format';

import type { QueryFilter } from '@/types';
import { Logger } from '@/logger';
import type { OAuth2Client, OAuth2ClientCreationInput } from '@/types';

import { backendRoutes } from '@/constants/routes';
import {
  defaultAPIRequestConfig,
  requestLogFunction,
} from '@/requests/defaults';

const logger = new Logger().withDebugValue(
  'source',
  'src/requests/oauth2clients.ts',
);

export function fetchListOfOAuth2Clients(
  qf: QueryFilter,
  adminMode: boolean = false,
): Promise<AxiosResponse> {
  const outboundURLParams = qf.toURLSearchParams();

  if (adminMode) {
    outboundURLParams.set('admin', 'true');
  }

  const uri = `${
    backendRoutes.GET_OAUTH2_CLIENTS
  }?${outboundURLParams.toString()}`;
  return axios
    .get(uri, defaultAPIRequestConfig)
    .then(requestLogFunction(logger, uri));
}

export function createOAuth2Client(
  oauth2client: OAuth2ClientCreationInput,
): Promise<AxiosResponse> {
  const uri = backendRoutes.CREATE_OAUTH2_CLIENT;
  return axios
    .post(uri, oauth2client, defaultAPIRequestConfig)
    .then(requestLogFunction(logger, uri));
}

export function fetchOAuth2Client(id: number): Promise<AxiosResponse> {
  const uri = format(backendRoutes.INDIVIDUAL_OAUTH2_CLIENT, id.toString());
  return axios
    .get(uri, defaultAPIRequestConfig)
    .then(requestLogFunction(logger, uri));
}

export function saveOAuth2Client(
  oauth2client: OAuth2Client,
): Promise<AxiosResponse> {
  const uri = format(
    backendRoutes.INDIVIDUAL_OAUTH2_CLIENT,
    oauth2client.id.toString(),
  );
  return axios
    .put(uri, oauth2client, defaultAPIRequestConfig)
    .then(requestLogFunction(logger, uri));
}

export function deleteOAuth2Client(id: number): Promise<AxiosResponse> {
  const uri = format(backendRoutes.INDIVIDUAL_OAUTH2_CLIENT, id.toString());
  return axios
    .delete(uri, defaultAPIRequestConfig)
    .then(requestLogFunction(logger, uri));
}

export function fetchAuditLogEntriesForOAuth2Client(
  id: number,
): Promise<AxiosResponse> {
  const uri = format(
    backendRoutes.INDIVIDUAL_OAUTH2_CLIENT_AUDIT_LOG,
    id.toString(),
  );
  return axios
    .get(uri, defaultAPIRequestConfig)
    .then(requestLogFunction(logger, uri));
}
