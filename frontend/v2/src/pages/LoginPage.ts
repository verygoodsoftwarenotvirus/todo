import axios, { AxiosResponse } from 'axios';
import { backendRoutes } from "../constants/routes";
import { AuthStatus, LoginRequest } from "../models";
import { methods } from "../constants/http";

export const getUserInfo = () =>
    axios.get( backendRoutes.USER_AUTH_STATUS, {
        method: methods.GET,
        withCredentials: true,
    });

export class LoginPage {
    constructor() {
    }

    static login = async function(creds: LoginRequest) {
        return axios.post(backendRoutes.LOGIN, creds, {withCredentials: true})
            .then(() => {
                getUserInfo().then((statusResponse: AxiosResponse<AuthStatus>) => {
                    return statusResponse;
                });
            })
            .catch(() => {
                return new AuthStatus();
            });
    }
}
