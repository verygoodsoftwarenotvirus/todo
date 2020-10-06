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
    usernameInput: string;
    passwordInput: string;
    totpTokenInput: string;

    constructor() {
        this.usernameInput = '';
        this.passwordInput = '';
        this.totpTokenInput = '';
    }

    state = () => {
        console.log(`
    username: ${this.usernameInput}
    password: ${this.passwordInput}
    2FA code: ${this.totpTokenInput}
        `)
    }

    buildLoginRequest = (): LoginRequest => {
        return {
            username: this.usernameInput,
            password: this.passwordInput,
            totpToken: this.totpTokenInput,
        } as LoginRequest
    }

    login = async () => {
        let creds: LoginRequest = this.buildLoginRequest();

        return axios.post(backendRoutes.LOGIN, creds, {withCredentials: true}).then(() => {
                getUserInfo().then((statusResponse: AxiosResponse<AuthStatus>) => {
                    return statusResponse;
                });
            })
            .catch((reason: object) => {
                console.error(`something went awry: ${reason}`)
                return new AuthStatus();
            });
    }
}
