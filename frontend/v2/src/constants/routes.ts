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

  // Items
  CREATE_ITEM = '/api/v1/items',
  GET_ITEMS = '/api/v1/items',
  INDIVIDUAL_ITEM = '/api/v1/items/{}',
  INDIVIDUAL_ITEM_AUDIT_LOG = '/api/v1/items/{}/audit',
  SEARCH_ITEMS = '/api/v1/items/search',
}
