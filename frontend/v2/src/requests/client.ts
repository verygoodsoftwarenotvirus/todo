import {
  login,
  logout,
  selfRequest,
  registrationRequest,
  passwordChangeRequest,
  checkAuthStatusRequest,
  validateTOTPSecretWithToken,
  twoFactorSecretChangeRequest,
} from './auth';

import {
  saveUser,
  fetchUser,
  deleteUser,
  fetchListOfUsers,
  fetchAuditLogEntriesForUser,
} from './users';

import {
  fetchListOfAuditLogEntries,
  fetchAuditLogEntry,
} from './audit_log_entries';

import {
  createOAuth2Client,
  fetchOAuth2Client,
  saveOAuth2Client,
  deleteOAuth2Client,
  fetchListOfOAuth2Clients,
  fetchAuditLogEntriesForOAuth2Client,
} from './oauth2clients';

import {
  saveWebhook,
  fetchWebhook,
  deleteWebhook,
  fetchListOfWebhooks,
  fetchAuditLogEntriesForWebhook,
} from './webhooks';

import {
  saveItem,
  fetchItem,
  createItem,
  deleteItem,
  searchForItems,
  fetchListOfItems,
  fetchAuditLogEntriesForItem,
} from './items';

import { createWebhook } from '@/requests/webhooks';
import { cycleCookieSecret } from './admin';

export class V1APIClient {
  // users stuff
  static fetchUser = fetchUser;
  static fetchListOfUsers = fetchListOfUsers;
  static saveUser = saveUser;
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

  // oauth2 clients stuff
  static createOAuth2Client = createOAuth2Client;
  static fetchOAuth2Client = fetchOAuth2Client;
  static saveOAuth2Client = saveOAuth2Client;
  static deleteOAuth2Client = deleteOAuth2Client;
  static fetchListOfOAuth2Clients = fetchListOfOAuth2Clients;
  static fetchAuditLogEntriesForOAuth2Client = fetchAuditLogEntriesForOAuth2Client;

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
