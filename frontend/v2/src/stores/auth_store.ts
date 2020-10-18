import axios, {AxiosError, AxiosResponse} from "axios";
import { writable } from 'svelte/store';

import type {UserStatus} from "@/models";

function buildAuthStatus() {
    const { subscribe, set, update } = writable({});

    const ass = {
        subscribe,
        setAuthStatus: (x: UserStatus) => set(x),
        logout: () => set({}),
    };

    axios.get("/auth/status", { withCredentials: true })
        .then((response: AxiosResponse<UserStatus>) => {
            ass.setAuthStatus(response.data);
        })
        .catch((err: AxiosError) => {
            console.error(err);
        })

    return ass;
}

export const authStatusStore = buildAuthStatus();