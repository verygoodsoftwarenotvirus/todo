import { backendRoutes } from '../constants/routes';
import { Logger } from '../logger';
import type {
  AuditLogEntry,
  QueryFilter,
  Webhook,
  WebhookCreationInput,
  WebhookList,
} from '../types';
import axios, { AxiosResponse } from 'axios';
import format from 'string-format';

const logger = new Logger().withDebugValue('source', 'src/apiClient/items.ts');

export function fetchListOfWebhooks(
  qf: QueryFilter,
  adminMode: boolean = false,
): Promise<AxiosResponse<WebhookList>> {
  const outboundURLParams = qf.toURLSearchParams();

  if (adminMode) {
    outboundURLParams.set('admin', 'true');
  }

  return axios.get(
    `${backendRoutes.GET_WEBHOOKS}?${outboundURLParams.toString()}`,
  );
}

export function createWebhook(
  input: WebhookCreationInput,
): Promise<AxiosResponse<Webhook>> {
  return axios.post(backendRoutes.CREATE_WEBHOOK, input);
}

export function fetchWebhook(id: number): Promise<AxiosResponse<Webhook>> {
  return axios.get(format(backendRoutes.INDIVIDUAL_WEBHOOK, id.toString()));
}

export function saveWebhook(webhook: Webhook): Promise<AxiosResponse<Webhook>> {
  return axios.put(
    format(backendRoutes.INDIVIDUAL_WEBHOOK, webhook.id.toString()),
    webhook,
  );
}

export function deleteWebhook(id: number): Promise<AxiosResponse> {
  return axios.delete(format(backendRoutes.INDIVIDUAL_WEBHOOK, id.toString()));
}

export function fetchAuditLogEntriesForWebhook(
  id: number,
): Promise<AxiosResponse<AuditLogEntry[]>> {
  return axios.get(
    format(backendRoutes.INDIVIDUAL_WEBHOOK_AUDIT_LOG, id.toString()),
  );
}
