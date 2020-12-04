import { backendRoutes } from '@/constants/routes';
import { Logger } from '@/logger';
import type {
  AuditLogEntry,
  Item,
  ItemCreationInput,
  ItemList,
  QueryFilter,
} from '@/types';
import axios, { AxiosResponse } from 'axios';
import format from 'string-format';

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

  const uri = `${backendRoutes.SEARCH_ITEMS}?${outboundURLParams.toString()}`;

  return axios.get(uri);
}

export function fetchListOfItems(
  qf: QueryFilter,
  adminMode: boolean = false,
): Promise<AxiosResponse<ItemList>> {
  const outboundURLParams = qf.toURLSearchParams();

  if (adminMode) {
    outboundURLParams.set('admin', 'true');
  }

  const uri = `${backendRoutes.GET_ITEMS}?${outboundURLParams.toString()}`;
  return axios.get(uri);
}

export function createItem(
  item: ItemCreationInput,
): Promise<AxiosResponse<Item>> {
  return axios.post(backendRoutes.CREATE_ITEM, item);
}

export function fetchItem(id: number): Promise<AxiosResponse<Item>> {
  const uri = format(backendRoutes.INDIVIDUAL_ITEM, id.toString());
  return axios.get(uri);
}

export function saveItem(item: Item): Promise<AxiosResponse<Item>> {
  const uri = format(backendRoutes.INDIVIDUAL_ITEM, item.id.toString());
  return axios.put(uri, item);
}

export function deleteItem(id: number): Promise<AxiosResponse> {
  const uri = format(backendRoutes.INDIVIDUAL_ITEM, id.toString());
  return axios.delete(uri);
}

export function fetchAuditLogEntriesForItem(
  id: number,
): Promise<AxiosResponse<AuditLogEntry[]>> {
  const uri = format(backendRoutes.INDIVIDUAL_ITEM_AUDIT_LOG, id.toString());
  return axios.get(uri);
}
