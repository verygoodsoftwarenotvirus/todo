import type { AxiosError, AxiosResponse } from 'axios';
import { writable } from 'svelte/store';

import { ErrorResponse, UserStatus } from '@/types';
import { Logger } from '@/logger';
import { V1APIClient } from '@/requests';

const logger = new Logger().withDebugValue(
  'source',
  'src/stores/user_status_store.ts',
);
const localStorageKey = 'userStatus';

function buildUserStatusStore() {
  const storedUserStatus: UserStatus = JSON.parse(
    localStorage.getItem(localStorageKey) || '{}',
  ) as UserStatus;
  const { subscribe, set } = writable(storedUserStatus);

  const userStatusStore = {
    subscribe,
    setUserStatus: (x: UserStatus) => {
      logger.withValue('userStatus', x).debug('setting user status');
      set(x);
    },
    logout: () => set(new UserStatus()),
  };

  userStatusStore.subscribe((value: UserStatus) => {
    localStorage.setItem(localStorageKey, JSON.stringify(value));
  });
  
  V1APIClient.checkAuthStatusRequest()
    .then((response: AxiosResponse<UserStatus>) => {
      userStatusStore.setUserStatus(response.data);
    })
    .catch((err: AxiosError<ErrorResponse>) => {
      logger.withValue('error', err).error('error checking for user status!');
      userStatusStore.setUserStatus(new UserStatus());
    });

  return userStatusStore;
}

export const userStatusStore = buildUserStatusStore();
