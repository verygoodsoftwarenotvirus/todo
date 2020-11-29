import { writable } from 'svelte/store';

const frontendOnlyMode =
  (process.env.FRONTEND_ONLY_MODE || '').toLowerCase() === 'true';

function createAdminModeStore() {
  const { subscribe, update } = writable<boolean>(frontendOnlyMode);

  return {
    subscribe,
    toggle: () => {
      update((n) => !n);
    },
  };
}

export const adminModeStore = createAdminModeStore();
