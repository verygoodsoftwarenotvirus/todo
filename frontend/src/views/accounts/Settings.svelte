<script lang="typescript">
import { onMount } from 'svelte';
import type { AxiosError, AxiosResponse } from 'axios';
import Select from 'svelte-select';

import {
  Account,
  ErrorResponse,
  UserStatus,
  UserSiteSettings,
  fakeAccountFactory,
} from '../../types';
import { Logger } from '../../logger';
import { V1APIClient } from '../../apiClient';
import { frontendRoutes } from '../../constants';
import { Superstore } from '../../stores';

export let accountID: number;
export let location: Location;

// removeme later
let memberships: object[] = [];
let members: object[] = [];

let logger = new Logger().withDebugValue(
  'source',
  'src/views/user/Settings.svelte',
);

let currentAuthStatus = new UserStatus();
let currentSessionSettings = new UserSiteSettings();
let translationsToUse = currentSessionSettings.getTranslations().pages.accountSettings;

let superstore = new Superstore({
  userStatusStoreUpdateFunc: (value: UserStatus) => {
    currentAuthStatus = value;
  },
  sessionSettingsStoreUpdateFunc: (value: UserSiteSettings) => {
    currentSessionSettings = value;
    translationsToUse = currentSessionSettings.getTranslations().pages.accountSettings;
  },
});

let ogUser: Account = new Account();
let account: Account = new Account();
let accountFetchError: string = '';


onMount(() => {
  if (superstore.frontendOnlyMode) {
    account = fakeAccountFactory.build();
    ogUser = { ...account };
  } else {
    V1APIClient.fetchAccount(accountID)
      .then((resp: AxiosResponse<Account>) => {
        account = resp.data;
        ogUser = { ...account };
      })
      .catch((err: AxiosError<ErrorResponse>) => {
        accountFetchError = err.message;
      });
  }
});
</script>

<div class="flex flex-wrap">
  <div class="w-full px-4">
    <div
      class="relative flex flex-col min-w-0 break-words w-full mb-6 shadow-lg rounded-lg bg-gray-200 border-0"
    >
      <div class="rounded-t bg-white mb-0 px-6 py-6">
        <div class="text-center flex justify-between">
          <h6 class="text-gray-800 text-xl font-bold">
            {translationsToUse.title}
          </h6>
        </div>
      </div>

      <div class="flex-auto px-4 lg:px-10 py-10 pt-0">
        <div class="text-center flex justify-between">
          <h6 class="text-gray-500 text-sm mt-3 mb-6 font-bold uppercase">
            {translationsToUse.sectionLabels.info}
          </h6>
        </div>
        <div class="flex flex-wrap">
          <div class="w-full lg:w-6/12 px-4">
            <div class="relative w-full mb-3">
              <label
                      class="block uppercase text-gray-700 text-xs font-bold mb-2"
                      for="grid-username"
              >
                {translationsToUse.inputLabels.name}
              </label>
              <input
                      id="grid-username"
                      type="text"
                      disabled
                      class="px-3 py-3 placeholder-gray-400 text-gray-700 bg-gray-300 rounded text-sm shadow focus:outline-none focus:shadow-outline w-full ease-linear transition-all duration-150"
                      bind:value="{account.name}"
              />
            </div>
          </div>
        </div>
        {#if currentAuthStatus.isAdmin()}
          <hr class="mt-6 border-b-1 border-gray-400" />

          <div class="text-center flex justify-between">
            <h6 class="text-gray-500 text-sm mt-3 mb-6 font-bold uppercase">
              {translationsToUse.sectionLabels.members}
            </h6>
            <button
              on:click="{console.log}"
              class="{false ? 'bg-blue-500' : 'bg-gray-300'} text-white active:bg-blue-600 font-bold uppercase text-xs rounded p-3 m-2"
              type="button"
            >
              {translationsToUse.buttons.saveMembers}
            </button>
          </div>
          <div class="flex flex-wrap">

            <div class="w-full lg:w-3/12 px-4">
              <div class="relative w-full mb-3">
                <ul>
                  <li>fart</li>
                  <li>fart</li>
                  <li>fart</li>
                  <li>fart</li>
                  <li>fart</li>
                </ul>
              </div>
            </div>
            <div class="w-full lg:w-3/12 px-4">
              <div class="relative w-full mb-3">
                <ul>
                  <li>fart</li>
                  <li>fart</li>
                  <li>fart</li>
                  <li>fart</li>
                  <li>fart</li>
                </ul>
              </div>
            </div>
            <div class="w-full lg:w-3/12 px-4">
              <div class="relative w-full mb-3">
                <ul>
                  <li>fart</li>
                  <li>fart</li>
                  <li>fart</li>
                  <li>fart</li>
                  <li>fart</li>
                </ul>
              </div>
            </div>
            <div class="w-full lg:w-3/12 px-4">
              <div class="relative w-full mb-3">
                <ul>
                  <li>fart</li>
                  <li>fart</li>
                  <li>fart</li>
                  <li>fart</li>
                  <li>fart</li>
                </ul>
              </div>
            </div>
          </div>
          {/if}
      </div>
    </div>
  </div>
</div>
