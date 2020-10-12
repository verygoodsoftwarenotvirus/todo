import { writable } from 'svelte/store';

import type {AuthStatus} from "@/models";

function buildAuthStatus() {
    const { subscribe, set, update } = writable({});

    return {
        subscribe,
        setAuthStatus: (authStatus: AuthStatus) => set(authStatus),
        logout: () => set({}),
    };
}

export const authStatus = buildAuthStatus();