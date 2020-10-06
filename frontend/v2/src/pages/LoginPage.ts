import axios, { AxiosResponse } from 'axios';
import { AuthStatus, LoginRequest } from "../models";

export const getUserInfo = () =>
    axios.get("/users/status", {withCredentials: true});

export class LoginPage {
    usernameInput: string;
    passwordInput: string;
    totpTokenInput: string;

    constructor() {
        this.usernameInput = '';
        this.passwordInput = '';
        this.totpTokenInput = '';
    }

    buildLoginRequest = (): LoginRequest => {
        return {
            username: this.usernameInput,
            password: this.passwordInput,
            totpToken: this.totpTokenInput,
        } as LoginRequest
    }

    login = async () => {
        const path = "/users/login"

        return axios.post(path, this.buildLoginRequest(), {withCredentials: true})
            .then(() => {
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
