import axios, { AxiosResponse } from "axios";
import format from "string-format";

import type { QueryFilter } from "@/types";
import { Logger } from "@/logger";
import type { Webhook, WebhookCreationInput, WebhookList } from "@/types";

import { backendRoutes } from "@/constants/routes";
import {
  defaultAPIRequestConfig,
  requestLogFunction,
} from "@/requests/defaults";

const logger = new Logger().withDebugValue("source", "src/requests/items.ts");

export function searchForWebhooks(
  query: string,
  qf: QueryFilter,
  adminMode: boolean = false
): Promise<AxiosResponse> {
  const outboundURLParams = qf.toURLSearchParams();
  if (adminMode) {
    outboundURLParams.set("admin", "true");
  }
  outboundURLParams.set("q", query);

  const uri = `${backendRoutes.SEARCH_ITEMS}?${outboundURLParams.toString()}`;

  return axios
    .get(uri, defaultAPIRequestConfig)
    .then(requestLogFunction(logger, uri));
}

export function fetchListOfWebhooks(
  qf: QueryFilter,
  adminMode: boolean = false
): Promise<AxiosResponse> {
  const outboundURLParams = qf.toURLSearchParams();

  if (adminMode) {
    outboundURLParams.set("admin", "true");
  }

  const uri = `${backendRoutes.GET_ITEMS}?${outboundURLParams.toString()}`;
  return axios
    .get(uri, defaultAPIRequestConfig)
    .then(requestLogFunction(logger, uri));
}

export function createWebhook(
  item: WebhookCreationInput
): Promise<AxiosResponse> {
  const uri = backendRoutes.CREATE_ITEM;
  return axios
    .post(uri, item, defaultAPIRequestConfig)
    .then(requestLogFunction(logger, uri));
}

export function fetchWebhook(id: number): Promise<AxiosResponse> {
  const uri = format(backendRoutes.INDIVIDUAL_ITEM, id.toString());
  return axios
    .get(uri, defaultAPIRequestConfig)
    .then(requestLogFunction(logger, uri));
}

export function saveWebhook(item: Webhook): Promise<AxiosResponse> {
  const uri = format(backendRoutes.INDIVIDUAL_ITEM, item.id.toString());
  return axios
    .put(uri, item, defaultAPIRequestConfig)
    .then(requestLogFunction(logger, uri));
}

export function deleteWebhook(id: number): Promise<AxiosResponse> {
  const uri = format(backendRoutes.INDIVIDUAL_ITEM, id.toString());
  return axios
    .delete(uri, defaultAPIRequestConfig)
    .then(requestLogFunction(logger, uri));
}

export function fetchAuditLogEntriesForWebhook(
  id: number
): Promise<AxiosResponse> {
  const uri = format(backendRoutes.INDIVIDUAL_ITEM_AUDIT_LOG, id.toString());
  return axios
    .get(uri, defaultAPIRequestConfig)
    .then(requestLogFunction(logger, uri));
}
