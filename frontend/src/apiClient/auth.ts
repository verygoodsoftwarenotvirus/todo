import { backendRoutes } from '../constants/routes';
import { Logger } from '../logger';
import type {
  LoginRequest,
  RegistrationRequest,
  TOTPTokenValidationRequest,
  User,
  UserPasswordUpdateRequest,
  UserRegistrationResponse,
  UserTwoFactorSecretUpdateRequest,
} from '../types';
import axios, { AxiosResponse } from 'axios';

const logger = new Logger().withDebugValue('source', 'src/apiClient/auth.ts');

export function checkAuthStatusRequest(): Promise<AxiosResponse> {
  return axios.get(backendRoutes.USER_AUTH_STATUS);
}

export function login(loginCreds: LoginRequest): Promise<AxiosResponse> {
  return axios.post(backendRoutes.LOGIN, loginCreds);
}

export function selfRequest(): Promise<AxiosResponse<User>> {
  return axios.get(backendRoutes.USER_SELF_INFO);
}

export function logout(): Promise<AxiosResponse> {
  return axios.post(backendRoutes.LOGOUT, {});
}

export function validateTOTPSecretWithToken(
  req: TOTPTokenValidationRequest,
): Promise<AxiosResponse> {
  return axios.post(backendRoutes.VERIFY_2FA_SECRET, req);
}

export function registrationRequest(
  req: RegistrationRequest,
): Promise<AxiosResponse<UserRegistrationResponse>> {
  return axios.post(backendRoutes.USER_REGISTRATION, req);
}

export function passwordChangeRequest(
  req: UserPasswordUpdateRequest,
): Promise<AxiosResponse> {
  return axios.put(backendRoutes.CHANGE_PASSWORD, req);
}

export function twoFactorSecretChangeRequest(
  req: UserTwoFactorSecretUpdateRequest,
): Promise<AxiosResponse> {
  return axios.post(backendRoutes.CHANGE_2FA_SECRET, req);
}
