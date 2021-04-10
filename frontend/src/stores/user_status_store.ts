import type { AxiosError, AxiosResponse } from 'axios';
import { writable } from 'svelte/store';
import { V1APIClient } from '../apiClient';
import { Logger } from '../logger';
import {
  AdminPermissionSummary,
  ErrorResponse,
  UserPermissionSummary,
  UserStatus,
} from '../types';

const logger = new Logger().withDebugValue(
  'source',
  'src/stores/user_status_store.ts',
);
const localStorageKey = 'userStatus';

const frontendOnlyMode =
  (process.env.FRONTEND_ONLY_MODE || '').toLowerCase() === 'true';

function buildUserStatusStore() {
  const storedUserStatus: UserStatus = JSON.parse(
    localStorage.getItem(localStorageKey) || '{}',
  ) as UserStatus;
  const { subscribe, set } = writable(storedUserStatus);

  const userStatusStore = {
    subscribe,
    setUserStatus: (x: UserStatus) => {
      set(x);
      logger
        .withDebugValue('userStatus', x as UserStatus)
        .debug('user status set');
    },
    logout: () => set(new UserStatus()),
  };

  userStatusStore.subscribe((value: UserStatus) => {
    localStorage.setItem(localStorageKey, JSON.stringify(value));
  });

  if (frontendOnlyMode) {
    const permMap = new Map<number, UserPermissionSummary>();
    userStatusStore.setUserStatus(
      new UserStatus(
        'good',
        'testing',
        true,
        permMap,
        new AdminPermissionSummary(true),
      ),
    );
  } else {
    V1APIClient.checkAuthStatusRequest()
      .then((response: AxiosResponse<UserStatus>) => {
        userStatusStore.setUserStatus(response.data);
      })
      .catch((err: AxiosError<ErrorResponse>) => {
        logger.withValue('error', err).error('error checking for user status!');
        userStatusStore.setUserStatus(new UserStatus());
      });
  }

  return userStatusStore;
}

export const userStatusStore = buildUserStatusStore();
