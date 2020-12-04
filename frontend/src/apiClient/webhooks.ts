import { backendRoutes } from '@/constants/routes';
import { Logger } from '@/logger';
import type {
  AuditLogEntry,
  QueryFilter,
  Webhook,
  WebhookCreationInput,
  WebhookList,
} from '@/types';
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

  const uri = `${backendRoutes.GET_ITEMS}?${outboundURLParams.toString()}`;
  return axios.get(uri);
}

export function createWebhook(
  item: WebhookCreationInput,
): Promise<AxiosResponse<Webhook>> {
  return axios.post(backendRoutes.CREATE_ITEM, item);
}

export function fetchWebhook(id: number): Promise<AxiosResponse<Webhook>> {
  const uri = format(backendRoutes.INDIVIDUAL_ITEM, id.toString());
  return axios.get(uri);
}

export function saveWebhook(item: Webhook): Promise<AxiosResponse<Webhook>> {
  const uri = format(backendRoutes.INDIVIDUAL_ITEM, item.id.toString());
  return axios.put(uri, item);
}

export function deleteWebhook(id: number): Promise<AxiosResponse> {
  const uri = format(backendRoutes.INDIVIDUAL_ITEM, id.toString());
  return axios.delete(uri);
}

export function fetchAuditLogEntriesForWebhook(
  id: number,
): Promise<AxiosResponse<AuditLogEntry[]>> {
  const uri = format(backendRoutes.INDIVIDUAL_ITEM_AUDIT_LOG, id.toString());
  return axios.get(uri);
}
