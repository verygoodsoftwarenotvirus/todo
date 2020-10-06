export enum backendRoutes {
  USER_REGISTRATION  = "/users/",
  USER_AUTH_STATUS   = "/users/status",
  LOGIN              = "/users/login",
  LOGOUT             = "/users/logout",
  VERIFY_2FA_SECRET  = "/users/totp_secret/verify",
  VALID_INGREDIENTS  = "/api/v1/valid_ingredients",
  VALID_INGREDIENT   = "/api/v1/valid_ingredients/{}",
  VALID_INSTRUMENTS  = "/api/v1/valid_instruments",
  VALID_INSTRUMENT   = "/api/v1/valid_instruments/{}",
  VALID_PREPARATIONS = "/api/v1/valid_preparations",
  VALID_PREPARATION  = "/api/v1/valid_preparations/{}",
}
