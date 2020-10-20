export enum backendRoutes {
  // User accounts
  USER_REGISTRATION   = "/users/",
  LOGIN               = "/users/login",
  LOGOUT              = "/users/logout",
  INDIVIDUAL_USER     = "/api/v1/users/{}",

  // Auth
  USER_AUTH_STATUS    = "/auth/status",
  USER_SELF_INFO      = "/api/v1/users/self",
  VERIFY_2FA_SECRET   = "/users/totp_secret/verify",

  // Items
  CREATE_ITEM         = "/api/v1/items",
  GET_ITEMS           = "/api/v1/items",
  INDIVIDUAL_ITEM     = "/api/v1/items/{}",
  SEARCH_ITEMS        = "/api/v1/items/search",
}
