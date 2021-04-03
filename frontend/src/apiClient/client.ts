import { Logger } from '@/logger';
import axios, { AxiosRequestConfig, AxiosResponse } from 'axios';
import {
  createAccount,
  deleteAccount,
  fetchAccount,
  fetchAuditLogEntriesForAccount,
  fetchListOfAccounts,
  saveAccount,
} from './accounts';
import { cycleCookieSecret } from './admin';
import {
  createAPIClient,
  deleteAPIClient,
  fetchAPIClient,
  fetchAuditLogEntriesForAPIClient,
  fetchListOfAPIClients,
} from './apiClients';
import {
  fetchAuditLogEntry,
  fetchListOfAuditLogEntries,
} from './audit_log_entries';
import {
  checkAuthStatusRequest,
  login,
  logout,
  passwordChangeRequest,
  registrationRequest,
  selfRequest,
  twoFactorSecretChangeRequest,
  validateTOTPSecretWithToken,
} from './auth';
import {
  createItem,
  deleteItem,
  fetchAuditLogEntriesForItem,
  fetchItem,
  fetchListOfItems,
  saveItem,
  searchForItems,
} from './items';
import {
  deleteUser,
  fetchAuditLogEntriesForUser,
  fetchListOfUsers,
  fetchUser,
} from './users';
import {
  createWebhook,
  deleteWebhook,
  fetchAuditLogEntriesForWebhook,
  fetchListOfWebhooks,
  fetchWebhook,
  saveWebhook,
} from './webhooks';

const logger = new Logger().withDebugValue('source', 'src/apiClient/client.ts');

axios.interceptors.request.use((request: AxiosRequestConfig):
  | AxiosRequestConfig
  | Promise<AxiosRequestConfig> => {
  logger
    .withDebugValue('url', request.url || '')
    .withDebugValue('method', request.method || '')
    .withDebugValue('body', request.data || null)
    .debug(`executing request`);

  request.withCredentials = true;

  return request;
});

axios.interceptors.response.use((response: AxiosResponse):
  | AxiosResponse
  | Promise<AxiosResponse> => {
  logger
    .withDebugValue('url', response.config.url || '')
    .withDebugValue('responseStatus', response?.status.toString())
    .withDebugValue('responseBody', JSON.stringify(response.data))
    .withDebugValue('requestBody', response.config.data || null)
    .debug(`response received`);

  return response;
});

export class V1APIClient {
  // users stuff
  static fetchUser = fetchUser;
  static fetchListOfUsers = fetchListOfUsers;
  static deleteUser = deleteUser;

  // auth stuff
  static login = login;
  static logout = logout;
  static selfRequest = selfRequest;
  static passwordChangeRequest = passwordChangeRequest;
  static twoFactorSecretChangeRequest = twoFactorSecretChangeRequest;
  static registrationRequest = registrationRequest;
  static checkAuthStatusRequest = checkAuthStatusRequest;
  static validateTOTPSecretWithToken = validateTOTPSecretWithToken;
  static fetchAuditLogEntriesForUser = fetchAuditLogEntriesForUser;

  // admin stuff
  static cycleCookieSecret = cycleCookieSecret;

  // audit log entries
  static fetchListOfAuditLogEntries = fetchListOfAuditLogEntries;
  static fetchAuditLogEntry = fetchAuditLogEntry;

  // accounts stuff
  static createAccount = createAccount;
  static fetchAccount = fetchAccount;
  static saveAccount = saveAccount;
  static deleteAccount = deleteAccount;
  static fetchListOfAccounts = fetchListOfAccounts;
  static fetchAuditLogEntriesForAccount = fetchAuditLogEntriesForAccount;

  // API clients stuff
  static createAPIClient = createAPIClient;
  static fetchAPIClient = fetchAPIClient;
  static deleteAPIClient = deleteAPIClient;
  static fetchListOfAPIClients = fetchListOfAPIClients;
  static fetchAuditLogEntriesForAPIClient = fetchAuditLogEntriesForAPIClient;

  // webhooks stuff
  static createWebhook = createWebhook;
  static fetchWebhook = fetchWebhook;
  static saveWebhook = saveWebhook;
  static deleteWebhook = deleteWebhook;
  static fetchListOfWebhooks = fetchListOfWebhooks;
  static fetchAuditLogEntriesForWebhook = fetchAuditLogEntriesForWebhook;

  // items stuff
  static createItem = createItem;
  static fetchItem = fetchItem;
  static saveItem = saveItem;
  static deleteItem = deleteItem;
  static searchForItems = searchForItems;
  static fetchListOfItems = fetchListOfItems;
  static fetchAuditLogEntriesForItem = fetchAuditLogEntriesForItem;
}
