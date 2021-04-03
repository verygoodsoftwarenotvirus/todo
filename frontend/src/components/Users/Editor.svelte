<script lang="typescript">
import { onMount } from 'svelte';
import type { AxiosError, AxiosResponse } from 'axios';

import {
  UserSiteSettings,
  User,
  UserStatus,
  AuditLogEntry,
  fakeUserFactory,
  fakeAuditLogEntryFactory,
} from '../../types';
import { Logger } from '../../logger';
import { V1APIClient } from '../../apiClient';
import AuditLogTable from '../AuditLogTable/AuditLogTable.svelte';
import { Superstore } from '../../stores';
import { renderUnixTime } from '../../utils';

export let location?: Location;
export let userID: number = 0;

// local state
let originalUser: User = new User();
let user: User = new User();

let needsToBeSaved: boolean = false;

let userRetrievalError: string = '';

function evaluateChanges() {
  needsToBeSaved = !User.areEqual(user, originalUser);
}

let logger = new Logger().withDebugValue(
  'source',
  'src/components/Editors/Account.svelte',
);

let adminMode: boolean = false;
let currentAuthStatus: UserStatus = new UserStatus();
let currentSessionSettings = new UserSiteSettings();
let translationsToUse = currentSessionSettings.getTranslations().models.user;

let superstore = new Superstore({
  userStatusStoreUpdateFunc: (value: UserStatus) => {
    currentAuthStatus = value;
  },
  sessionSettingsStoreUpdateFunc: (value: UserSiteSettings) => {
    currentSessionSettings = value;
    translationsToUse = currentSessionSettings.getTranslations().models.user;
  },
  adminModeUpdateFunc: (value: boolean) => {
    adminMode = value;
  },
});

function fetchUser(): void {
  logger.debug(`fetchUser called`);

  if (userID === 0) {
    throw new Error('id cannot be zero!');
  }

  if (superstore.frontendOnlyMode) {
    user = fakeUserFactory.build();
  } else {
    V1APIClient.fetchUser(userID)
      .then((response: AxiosResponse<User>) => {
        user = { ...response.data };
        originalUser = { ...response.data };
      })
      .catch((error: AxiosError) => {
        userRetrievalError = error.response?.data;
      });
  }
}

function fetchAuditLogEntries(): Promise<AxiosResponse<AuditLogEntry[]>> {
  logger.debug(`fetchAuditLogEntries called`);

  if (userID === 0) {
    throw new Error('id cannot be zero!');
  }

  if (!adminMode) {
    return new Promise<AxiosResponse<AuditLogEntry[]>>((resolve) => {
      resolve({ data: [] } as AxiosResponse);
    });
  }

  if (superstore.frontendOnlyMode) {
    return new Promise<AxiosResponse<AuditLogEntry[]>>((resolve) => {
      resolve({ data: fakeAuditLogEntryFactory.buildList(10) } as AxiosResponse);
    });
  } else {
    return V1APIClient.fetchAuditLogEntriesForUser(userID);
  }
}

onMount(fetchUser);
</script>

<div
  class="relative flex flex-col min-w-0 break-words bg-white w-full mb-6 shadow-lg rounded"
>
  <div class="rounded-t mb-0 px-4 py-3 bg-transparent justify-between ">
    <div class="flex flex-wrap items-center">
      <div class="relative w-full max-w-full flex-grow flex-1">
        {#if originalUser.id !== 0}
          <h2 class="text-gray-800 text-xl font-semibold">
            #{originalUser.id}:
            {originalUser.username}
          </h2>
        {/if}
      </div>
    </div>
  </div>
  <div>
    <div class="flex flex-wrap -mx-3 mb-6">
      <div class="w-full md:w-1/2 px-3 mb-6 md:mb-0">
        <label
          class="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2"
          for="grid-username"
        >
          {translationsToUse.labels.username}
        </label>
        <input
          class="appearance-none block w-full text-gray-700 border border-red-500 rounded py-3 px-4 mb-3 leading-tight focus:outline-none focus:bg-white"
          id="grid-username"
          type="text"
          disabled
          placeholder="{translationsToUse.inputPlaceholders.username}"
          on:keyup="{evaluateChanges}"
          bind:value="{user.username}"
        />

        <div class="m-3 inline-flex">
          <label
            class="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2"
            for="grid-is-admin"
          >
            {translationsToUse.labels.isAdmin}: &nbsp;
          </label>
          <div id="grid-is-admin">{user.serviceAdminPermissions !== 0 ? 'yes' : 'no'}</div>
        </div>

        <div></div>

        <div class="m-3 inline-flex">
          <label
            class="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2"
            for="grid-password-last-changed-on"
          >
            {translationsToUse.labels.passwordLastChangedOn}: &nbsp;
          </label>
          <div id="grid-password-last-changed-on">
            {renderUnixTime(user.passwordLastChangedOn)}
          </div>
        </div>

        <div></div>

        <div class="m-3 inline-flex">
          <label
            class="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2"
            for="grid-created-on"
          >
            {translationsToUse.labels.createdOn}: &nbsp;
          </label>
          <p id="grid-created-on">{renderUnixTime(user.createdOn)}</p>
        </div>

        <div></div>

        <div class="m-3 inline-flex">
          <label
            class="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2"
            for="grid-last-updated-on"
          >
            {translationsToUse.labels.lastUpdatedOn}: &nbsp;
          </label>
          <p id="grid-last-updated-on">{renderUnixTime(user.lastUpdatedOn)}</p>
        </div>
      </div>

      {#if currentAuthStatus.isAdmin() && adminMode}
        <div
          class="flex w-full mr-3 mt-4 max-w-full flex-grow justify-end flex-1"
        >
          <button
            class="bg-red-500 hover:bg-red-700 text-white font-bold py-2 px-4 rounded"
          ><i class="fa fa-trash-alt"></i>
            {translationsToUse.actions.ban}</button>
        </div>
      {/if}
    </div>
  </div>

  {#if currentAuthStatus.isAdmin() && adminMode}
    <AuditLogTable entryFetchFunc="{fetchAuditLogEntries}" />
  {/if}
</div>
