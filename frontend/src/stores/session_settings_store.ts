import { Logger } from '@/logger';
import { UserSiteSettings } from '@/types';
import { writable } from 'svelte/store';

const logger = new Logger().withDebugValue(
  'source',
  'src/stores/session_settings_store.ts',
);

function buildSessionSettingsStore() {
  const { subscribe, set } = writable(new UserSiteSettings());

  const sessionSettingsStore = {
    subscribe,
    updateSettings: (x: UserSiteSettings) => {
      logger.withValue('settings', x).debug('session settings updated');
      set(x);
    },
  };

  sessionSettingsStore.updateSettings(new UserSiteSettings());

  return sessionSettingsStore;
}

export const sessionSettingsStore = buildSessionSettingsStore();
