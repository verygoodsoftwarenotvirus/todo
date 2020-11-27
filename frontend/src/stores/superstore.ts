import type { UserStatus, UserSiteSettings } from '@/types';
import { userStatusStore } from './user_status_store';
import { sessionSettingsStore } from './session_settings_store';
import { adminModeStore } from './admin_mode_store';

interface functionSet {
  userStatusStoreUpdateFunc?: (value: UserStatus) => void;
  sessionSettingsStoreUpdateFunc?: (value: UserSiteSettings) => void;
  adminModeUpdateFunc?: (value: boolean) => void;
}

export class Superstore {
  userStatusStoreUpdateFunc?: (value: UserStatus) => void;
  sessionSettingsStoreUpdateFunc?: (value: UserSiteSettings) => void;
  adminModeUpdateFunc?: (value: boolean) => void;

  constructor({
    userStatusStoreUpdateFunc,
    sessionSettingsStoreUpdateFunc,
    adminModeUpdateFunc,
  }: functionSet) {
    this.userStatusStoreUpdateFunc = userStatusStoreUpdateFunc;
    this.sessionSettingsStoreUpdateFunc = sessionSettingsStoreUpdateFunc;
    this.adminModeUpdateFunc = adminModeUpdateFunc;

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

  toggleAdminMode() {
    adminModeStore.toggle();
  }
}
