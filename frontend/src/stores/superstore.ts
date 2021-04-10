import { Logger } from '../logger';
import type { UserSiteSettings, UserStatus } from '../types';
import { adminModeStore } from './admin_mode_store';
import { sessionSettingsStore } from './session_settings_store';
import { userStatusStore } from './user_status_store';

interface functionSet {
  userStatusStoreUpdateFunc?: (value: UserStatus) => void;
  sessionSettingsStoreUpdateFunc?: (value: UserSiteSettings) => void;
  adminModeUpdateFunc?: (value: boolean) => void;
}

let logger = new Logger().withDebugValue('source', 'src/stores/superstore.ts');

const frontendOnlyMode =
  (process.env.FRONTEND_ONLY_MODE || '').toLowerCase() === 'true';

export class Superstore {
  userStatusStoreUpdateFunc?: (value: UserStatus) => void;
  sessionSettingsStoreUpdateFunc?: (value: UserSiteSettings) => void;
  adminModeUpdateFunc?: (value: boolean) => void;

  frontendOnlyMode: boolean;

  constructor({
    userStatusStoreUpdateFunc,
    sessionSettingsStoreUpdateFunc,
    adminModeUpdateFunc,
  }: functionSet) {
    this.userStatusStoreUpdateFunc = userStatusStoreUpdateFunc;
    this.sessionSettingsStoreUpdateFunc = sessionSettingsStoreUpdateFunc;
    this.adminModeUpdateFunc = adminModeUpdateFunc;

    this.frontendOnlyMode = frontendOnlyMode;

    if (this.userStatusStoreUpdateFunc) {
      userStatusStore.subscribe(this.userStatusStoreUpdateFunc);
    }

    if (this.sessionSettingsStoreUpdateFunc) {
      sessionSettingsStore.subscribe(this.sessionSettingsStoreUpdateFunc);
    }

    if (this.adminModeUpdateFunc) {
      adminModeStore.subscribe(this.adminModeUpdateFunc);
    }
  }

  static frontendOnlyMode: boolean = frontendOnlyMode;

  setUserStatus(x: UserStatus) {
    userStatusStore.setUserStatus(x);
  }

  toggleAdminMode() {
    adminModeStore.toggle();
  }
}
