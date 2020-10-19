import axios, { AxiosError, AxiosResponse } from "axios";
import { writable } from 'svelte/store';

import { UserStatus } from "@/models";
import { Logger } from "@/logger";

const logger = new Logger().withDebugValue("source", "src/stores/auth_store.ts");

function buildAuthStatus() {
    const { subscribe, set, update } = writable({});

    const ass = {
        subscribe,
        setAuthStatus: (x: UserStatus) => {
            logger.withValue("userStatus", x).debug("setting auth status");
            set(x);
        },
        logout: () => set({}),
    };

    axios.get("/auth/status", { withCredentials: true })
        .then((response: AxiosResponse<UserStatus>) => {
            ass.setAuthStatus(response.data);
        })
        .catch((err: AxiosError) => {
            ass.setAuthStatus(new UserStatus());
            console.error(err);
        })

    const x = new UserStatus()
    x.isAuthenticated = true;
    ass.setAuthStatus(x)

    return ass;
}

export const authStatusStore = buildAuthStatus();