import { writable } from 'svelte/store';

import type {AuthStatus} from "@/models";

function buildAuthStatus() {
    const { subscribe, set, update } = writable({});

    return {
        subscribe,
        setAuthStatus: (x: AuthStatus) => set(x),
        logout: () => set({}),
    };
}

export const authStatusStore = buildAuthStatus();