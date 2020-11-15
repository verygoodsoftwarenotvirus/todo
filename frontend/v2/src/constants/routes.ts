export enum backendRoutes {
  // User accounts
  USER_REGISTRATION = '/users/',
  LOGIN = '/users/login',
  LOGOUT = '/users/logout',
  INDIVIDUAL_USER = '/api/v1/users/{}',
  INDIVIDUAL_USER_AUDIT_LOG = '/api/v1/users/{}/audit',

  // Auth
  USER_AUTH_STATUS = '/auth/status',
  USER_SELF_INFO = '/api/v1/users/self',
  CHANGE_PASSWORD = '/users/password/new',
  CHANGE_2FA_SECRET = '/users/totp_secret/new',
  VERIFY_2FA_SECRET = '/users/totp_secret/verify',

  // Admin
  CYCLE_COOKIE_SECRET = '/_admin_/cycle_cookie_secret',

  // Audit Log Entries
  GET_AUDIT_LOG_ENTRIES = '/_admin_/audit_log',
  INDIVIDUAL_AUDIT_LOG_ENTRY = '/_admin_/audit_log/{}',

  // OAuth2 Clients
  CREATE_OAUTH2_CLIENT = '/oauth2/clients',
  GET_OAUTH2_CLIENTS = '/api/v1/oauth2/clients',
  INDIVIDUAL_OAUTH2_CLIENT = '/api/v1/oauth2/clients/{}',
  INDIVIDUAL_OAUTH2_CLIENT_AUDIT_LOG = '/api/v1/oauth2/clients/{}/audit',

  // Items
  CREATE_ITEM = '/api/v1/items',
  GET_ITEMS = '/api/v1/items',
  INDIVIDUAL_ITEM = '/api/v1/items/{}',
  INDIVIDUAL_ITEM_AUDIT_LOG = '/api/v1/items/{}/audit',
  SEARCH_ITEMS = '/api/v1/items/search',
}

export enum frontendRoutes {
  // Pages
  LANDING = '/',
  ADMIN_DASHBOARD = '/admin/dashboard',

  // User accounts
  LOGIN = '/auth/login',
  USER_SETTINGS = '/user/settings',

  // OAuth2 Clients
  LIST_OAUTH2_CLIENTS = '/things/oauth2_clients',

  // Webhooks
  LIST_WEBHOOKS = '/things/webhooks',

  // Items
  LIST_ITEMS = '/things/items',
  INDIVIDUAL_ITEM = '/things/items/{}',
}
