import axios, { AxiosResponse } from "axios";

import type {
    RegistrationRequest,
    LoginRequest,
    TOTPTokenValidationRequest,
    UserPasswordUpdateRequest,
    UserTwoFactorSecretUpdateRequest,
} from "@/types";

import { Logger } from "@/logger";
import {backendRoutes} from "@/constants/routes";
import {defaultAPIRequestConfig, requestLogFunction} from "@/requests/defaults";

const logger = new Logger().withDebugValue("source", "src/requests/auth.ts");

export function checkAuthStatusRequest(): Promise<AxiosResponse> {
    const uri = backendRoutes.USER_AUTH_STATUS;
    return axios.get(uri, defaultAPIRequestConfig)
        .then(requestLogFunction(logger, uri));
}

export function login(loginCreds: LoginRequest): Promise<AxiosResponse> {
    const uri = backendRoutes.LOGIN;
    return axios.post(uri, loginCreds, defaultAPIRequestConfig)
        .then(requestLogFunction(logger, uri));
}

export function selfRequest(): Promise<AxiosResponse> {
    const uri = backendRoutes.USER_SELF_INFO;
    return axios.get(uri, defaultAPIRequestConfig)
        .then(requestLogFunction(logger, uri));
}

export function logout(): Promise<AxiosResponse> {
  const uri = backendRoutes.LOGOUT;
    return axios.post(uri, {}, defaultAPIRequestConfig)
      .then(requestLogFunction(logger, uri));
}

export function validateTOTPSecretWithToken(validationRequest: TOTPTokenValidationRequest): Promise<AxiosResponse> {
    const uri = backendRoutes.VERIFY_2FA_SECRET;
    return axios.post(uri, validationRequest, defaultAPIRequestConfig)
        .then(requestLogFunction(logger, uri));
}

export function registrationRequest(rr: RegistrationRequest): Promise<AxiosResponse> {
    const uri = backendRoutes.USER_REGISTRATION;
    return axios.post(uri, rr, defaultAPIRequestConfig)
        .then(requestLogFunction(logger, uri));
}

export function passwordChangeRequest(pcr: UserPasswordUpdateRequest): Promise<AxiosResponse> {
    const uri = backendRoutes.CHANGE_PASSWORD;
    return axios.put(uri, pcr, defaultAPIRequestConfig)
        .then(requestLogFunction(logger, uri));
}

export function twoFactorSecretChangeRequest(sur: UserTwoFactorSecretUpdateRequest): Promise<AxiosResponse> {
    const uri = backendRoutes.CHANGE_2FA_SECRET;
    return axios.post(uri, sur, defaultAPIRequestConfig)
        .then(requestLogFunction(logger, uri));
}