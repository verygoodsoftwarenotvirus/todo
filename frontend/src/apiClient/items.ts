import axios, { AxiosResponse } from 'axios';
import format from 'string-format';
import { backendRoutes } from '../constants/routes';
import { Logger } from '../logger';
import type {
  AuditLogEntry,
  Item,
  ItemCreationInput,
  ItemList,
  QueryFilter,
} from '../types';

const logger = new Logger().withDebugValue('source', 'src/apiClient/items.ts');

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

  return axios.get(
    `${backendRoutes.SEARCH_ITEMS}?${outboundURLParams.toString()}`,
  );
}

export function fetchListOfItems(
  qf: QueryFilter,
  adminMode: boolean = false,
): Promise<AxiosResponse<ItemList>> {
  const outboundURLParams = qf.toURLSearchParams();

  if (adminMode) {
    outboundURLParams.set('admin', 'true');
  }

  return axios.get(
    `${backendRoutes.GET_ITEMS}?${outboundURLParams.toString()}`,
  );
}

export function createItem(
  item: ItemCreationInput,
): Promise<AxiosResponse<Item>> {
  return axios.post(backendRoutes.CREATE_ITEM, item);
}

export function fetchItem(id: number): Promise<AxiosResponse<Item>> {
  return axios.get(format(backendRoutes.INDIVIDUAL_ITEM, id.toString()));
}

export function saveItem(item: Item): Promise<AxiosResponse<Item>> {
  return axios.put(
    format(backendRoutes.INDIVIDUAL_ITEM, item.id.toString()),
    item,
  );
}

export function deleteItem(id: number): Promise<AxiosResponse> {
  return axios.delete(format(backendRoutes.INDIVIDUAL_ITEM, id.toString()));
}

export function fetchAuditLogEntriesForItem(
  id: number,
): Promise<AxiosResponse<AuditLogEntry[]>> {
  return axios.get(
    format(backendRoutes.INDIVIDUAL_ITEM_AUDIT_LOG, id.toString()),
  );
}
