import axios, { AxiosResponse } from 'axios';
import format from 'string-format';

import type { QueryFilter } from '@/types';
import { Logger } from '@/logger';
import type { Item, ItemCreationInput, ItemList } from '@/types';

import { backendRoutes } from '@/constants/routes';
import {
  defaultAPIRequestConfig,
  requestLogFunction,
} from '@/requests/defaults';

const logger = new Logger().withDebugValue('source', 'src/requests/items.ts');

export function searchForItems(
  query: string,
  qf: QueryFilter,
  adminMode: boolean = false,
): Promise<AxiosResponse> {
  const outboundURLParams = qf.toURLSearchParams();
  if (adminMode) {
    outboundURLParams.set('admin', 'true');
  }
  outboundURLParams.set('q', query);

  const uri = `${backendRoutes.SEARCH_ITEMS}?${outboundURLParams.toString()}`;

  return axios
    .get(uri, defaultAPIRequestConfig)
    .then(requestLogFunction(logger, uri));
}

export function fetchListOfItems(
  qf: QueryFilter,
  adminMode: boolean = false,
): Promise<AxiosResponse> {
  const outboundURLParams = qf.toURLSearchParams();

  if (adminMode) {
    outboundURLParams.set('admin', 'true');
  }

  const uri = `${backendRoutes.GET_ITEMS}?${outboundURLParams.toString()}`;
  return axios
    .get(uri, defaultAPIRequestConfig)
    .then(requestLogFunction(logger, uri));
}

export function createItem(item: ItemCreationInput): Promise<AxiosResponse> {
  const uri = backendRoutes.CREATE_ITEM;
  return axios
    .post(uri, item, defaultAPIRequestConfig)
    .then(requestLogFunction(logger, uri));
}

export function fetchItem(id: number): Promise<AxiosResponse> {
  const uri = format(backendRoutes.INDIVIDUAL_ITEM, id.toString());
  return axios
    .get(uri, defaultAPIRequestConfig)
    .then(requestLogFunction(logger, uri));
}

export function saveItem(item: Item): Promise<AxiosResponse> {
  const uri = format(backendRoutes.INDIVIDUAL_ITEM, item.id.toString());
  return axios
    .put(uri, item, defaultAPIRequestConfig)
    .then(requestLogFunction(logger, uri));
}

export function deleteItem(id: number): Promise<AxiosResponse> {
  const uri = format(backendRoutes.INDIVIDUAL_ITEM, id.toString());
  return axios
    .delete(uri, defaultAPIRequestConfig)
    .then(requestLogFunction(logger, uri));
}

export function fetchAuditLogEntriesForItem(
  id: number,
): Promise<AxiosResponse> {
  const uri = format(backendRoutes.INDIVIDUAL_ITEM_AUDIT_LOG, id.toString());
  return axios
    .get(uri, defaultAPIRequestConfig)
    .then(requestLogFunction(logger, uri));
}
