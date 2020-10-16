import axios, {AxiosError, AxiosResponse} from "axios";
import { writable } from 'svelte/store';

import type {AuthStatus} from "@/models";

function buildAuthStatus() {
    const { subscribe, set, update } = writable({});

    const ass = {
        subscribe,
        setAuthStatus: (x: AuthStatus) => set(x),
        logout: () => set({}),
    };

    axios.get("/users/status", { withCredentials: true })
        .then((response: AxiosResponse<AuthStatus>) => {
            ass.setAuthStatus(response.data);
        })
        .catch((err: AxiosError) => {
            console.error(err);
        })

    return ass;
}

export const authStatusStore = buildAuthStatus();