import { writable } from 'svelte/store';

function buildAuthStatus() {
    const { subscribe, set, update } = writable({});

    return {
        subscribe,
        setAuthStatus: (authStatus) => set(authStatus),
        logout: () => set({}),
    };
}

export const authStatus = buildAuthStatus();