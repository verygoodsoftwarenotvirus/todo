import { writable } from 'svelte/store';

function buildAdminModeStatus() {
    const { subscribe, set, update } = writable({});

    let adminMode: boolean = false;

    return {
        subscribe,
        toggleAdminStatus: () => {
            adminMode = !adminMode
            set(adminMode);
        },
    };
}

export const adminModeStatus = buildAdminModeStatus();