import axios, {AxiosError, AxiosResponse} from 'axios';
import { navigate } from "svelte-routing";

import type { RegistrationRequest, UserRegistrationResponse, ErrorResponse } from "@/models";

const exampleQRCode = "data:image/jpeg;base64,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"

export class RegistrationPage {
    registrationMayProceed: boolean;
    usernameInput: string;
    passwordInput: string;
    passwordRepeatInput: string;

    constructor() {
        this.registrationMayProceed = false;
        this.usernameInput = '';
        this.passwordInput = '';
        this.passwordRepeatInput = '';
    }

    evaluateInputs = (): void => {
        this.registrationMayProceed = (
            this.usernameInput !== "" &&
            this.passwordInput !== "" &&
            this.passwordRepeatInput !== "" &&
            this.passwordInput === this.passwordRepeatInput
        )
        console.debug(`evaluateInputs called, this.registrationMayProceed = ${this.registrationMayProceed}`);
    }

    buildRegistrationRequest = (): RegistrationRequest => {
        return {
            username: this.usernameInput,
            password: this.passwordInput,
            repeatedPassword: this.passwordRepeatInput,
        } as RegistrationRequest;
    }

    register = async () => {
        console.debug("RegistrationPage.register called")

        const path = "/users/"

        if (!this.registrationMayProceed) {
            // this should never occur
            throw new Error("registration input is not valid!");
        }

        return axios.post(path, this.buildRegistrationRequest())
            .then((response: AxiosResponse<UserRegistrationResponse | ErrorResponse>) => {
                const data = response.data as UserRegistrationResponse;
                console.dir(data);
                return data;
            })
            .catch((reason: AxiosError) => {
                if (reason.response) {
                    const data = reason.response.data as ErrorResponse;
                    console.error(data.message);
                }
            });
    }
}
