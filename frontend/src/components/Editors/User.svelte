<script lang="typescript">
import { AxiosError, AxiosResponse } from 'axios';

import { UserSiteSettings, User, UserStatus, AuditLogEntry } from '../../types';
import { Logger } from '../../logger';
import { V1APIClient } from '../../apiClient';
import AuditLogTable from '../AuditLogTable/AuditLogTable.svelte';
import { Superstore } from '../../stores/superstore';

export let location: Location;
export let userID: number = 0;

// local state
let originalUser: User = new User();
let user: User = new User();
let auditLogEntries: AuditLogEntry[] = [];

let needsToBeSaved: boolean = false;
let userRetrievalError: string = '';

function evaluateChanges() {
  needsToBeSaved = !User.areEqual(user, originalUser);
}

let logger = new Logger().withDebugValue(
  'source',
  'src/components/Editors/User.svelte',
);

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
});

function fetchUser(): void {
  logger.debug(`fetchUser called`);

  if (userID === 0) {
    throw new Error('id cannot be zero!');
  }

  V1APIClient.fetchUser(userID)
    .then((response: AxiosResponse<User>) => {
      user = { ...response.data };
      originalUser = { ...response.data };
    })
    .catch((error: AxiosError) => {
      if (error.response && error.response.data) {
        userRetrievalError = error.response.data;
      }
    });
}
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
          for="grid-first-name"
        >
          {translationsToUse.myAccount}
        </label>
        <input
          class="appearance-none block w-full text-gray-700 border border-red-500 rounded py-3 px-4 mb-3 leading-tight focus:outline-none focus:bg-white"
          id="grid-first-name"
          type="text"
          placeholder="{translationsToUse.myAccount}"
          on:keyup="{evaluateChanges}"
          bind:value="{user.username}"
        />
      </div>
      <div
        class="flex w-full mr-3 mt-4 max-w-full flex-grow justify-end flex-1"
      >
        <button
          class="bg-red-500 hover:bg-red-700 text-white font-bold py-2 px-4 rounded"
          on:click="{console.log}"
        ><i class="fa fa-trash-alt"></i>
          {translationsToUse.myAccount}</button>
      </div>
    </div>
  </div>

  {#if currentUserStatus.isAdmin}
    <AuditLogTable entries="{auditLogEntries}" />
  {/if}
</div>
