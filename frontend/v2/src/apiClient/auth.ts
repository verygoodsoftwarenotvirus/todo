import axios, { AxiosResponse } from 'axios';

import type {
  RegistrationRequest,
  LoginRequest,
  TOTPTokenValidationRequest,
  User,
  UserRegistrationResponse,
  UserPasswordUpdateRequest,
  UserTwoFactorSecretUpdateRequest,
} from '@/types';

import { Logger } from '@/logger';
import { backendRoutes } from '@/constants/routes';

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
  validationRequest: TOTPTokenValidationRequest,
): Promise<AxiosResponse> {
  return axios.post(backendRoutes.VERIFY_2FA_SECRET, validationRequest);
}

export function registrationRequest(
  rr: RegistrationRequest,
): Promise<AxiosResponse<UserRegistrationResponse>> {
  return axios.post(backendRoutes.USER_REGISTRATION, rr);
}

export function passwordChangeRequest(
  pcr: UserPasswordUpdateRequest,
): Promise<AxiosResponse> {
  return axios.put(backendRoutes.CHANGE_PASSWORD, pcr);
}

export function twoFactorSecretChangeRequest(
  sur: UserTwoFactorSecretUpdateRequest,
): Promise<AxiosResponse> {
  return axios.post(backendRoutes.CHANGE_2FA_SECRET, sur);
}
