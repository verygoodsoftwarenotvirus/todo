import { writable } from 'svelte/store';

import {SessionSettings} from "@/models";
import { Logger } from "@/logger";

const logger = new Logger().withDebugValue("source", "src/stores/session_settings_store.ts");

function buildSessionSettingsStore() {
    const { subscribe, set } = writable({});

    const sessionSettingsStore = {
        subscribe,
        updateSettings: (x: SessionSettings) => {
            logger.withValue("settings", x).debug("site settings updated");
            set(x);
        },
    };

    sessionSettingsStore.updateSettings(new SessionSettings())

    return sessionSettingsStore;
}

export const sessionSettingsStore = buildSessionSettingsStore();