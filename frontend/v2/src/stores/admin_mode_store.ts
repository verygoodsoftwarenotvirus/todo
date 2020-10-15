import { writable } from 'svelte/store';

function createAdminModeStore() {
    const { subscribe, update } = writable<boolean>(false);

    return {
        subscribe,
        toggle: () => update(n => !n),
    };
}

export const adminModeStore = createAdminModeStore();