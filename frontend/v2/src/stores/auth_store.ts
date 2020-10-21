import type { AxiosError, AxiosResponse } from "axios";
import { writable } from 'svelte/store';

import {ErrorResponse, UserStatus} from "@/models";
import { Logger } from "@/logger";
import { V1APIClient } from "@/requests";

const logger = new Logger().withDebugValue("source", "src/stores/auth_store.ts");

function buildUserStatusStore() {
    const { subscribe, set, update } = writable({});

    const userStatusStore = {
        subscribe,
        setAuthStatus: (x: UserStatus) => {
            logger.withValue("userStatus", x).debug("setting auth status");
            set(x);
        },
        logout: () => set(new UserStatus()),
    };

    V1APIClient.checkAuthStatusRequest()
        .then((response: AxiosResponse<UserStatus>) => {
            userStatusStore.setAuthStatus(response.data);
        })
        .catch((err: AxiosError<ErrorResponse>) => {
            userStatusStore.setAuthStatus(new UserStatus());
        });

    return userStatusStore;
}

export const userStatusStore = buildUserStatusStore();