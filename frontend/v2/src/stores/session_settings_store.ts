import { writable } from 'svelte/store';

import {UserSiteSettings} from "@/types";
import { Logger } from "@/logger";

const logger = new Logger().withDebugValue("source", "src/stores/session_settings_store.ts");

function buildSessionSettingsStore() {
    const { subscribe, set } = writable({});

    const sessionSettingsStore = {
        subscribe,
        updateSettings: (x: UserSiteSettings) => {
            logger.withValue("settings", x).debug("session settings updated");
            set(x);
        },
    };

    sessionSettingsStore.updateSettings(new UserSiteSettings());

    return sessionSettingsStore;
}

export const sessionSettingsStore = buildSessionSettingsStore();