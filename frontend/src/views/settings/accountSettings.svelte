<script lang="typescript">
import type { AxiosError, AxiosResponse } from 'axios';

import {
  Account,
  ErrorResponse,
  UserStatus,
  UserSiteSettings,
  fakeAccountFactory,
  AccountUserMembership,
} from '../../types';
import { Logger } from '../../logger';
import { V1APIClient } from '../../apiClient';
import { Superstore } from '../../stores';

export let accountID: number;
export let location: Location;

// TODO: remove me later
let memberships: object[] = [];
let members: object[] = [];

let logger = new Logger().withDebugValue(
  'source',
  'src/views/user/userSettings.svelte',
);

let frontendOnlyMode: boolean = false;
let currentAuthStatus = new UserStatus();
let currentSessionSettings = new UserSiteSettings();
let translationsToUse = currentSessionSettings.getTranslations();
let pageTranslations = translationsToUse.pages.accountSettings;
let modelTranslations = translationsToUse.models.accountUserMembership;

let superstore = new Superstore({
  userStatusStoreUpdateFunc: (value: UserStatus) => {
    currentAuthStatus = value;
    if (accountID !== currentAuthStatus.activeAccount) {
      accountID = currentAuthStatus.activeAccount;
      fetchAccount();
    }
  },
  sessionSettingsStoreUpdateFunc: (value: UserSiteSettings) => {
    currentSessionSettings = value;
    translationsToUse = currentSessionSettings.getTranslations();
    pageTranslations = translationsToUse.pages.accountSettings;
    modelTranslations = translationsToUse.models.accountUserMembership;
  },
});
frontendOnlyMode = superstore.frontendOnlyMode;

let adminMode: boolean = false;

let originalAccount: Account = new Account();
let account: Account = new Account();
let accountFetchError: string = '';

function fetchAccount() {
  if (frontendOnlyMode) {
    account = fakeAccountFactory.build();
    originalAccount = { ...account };
  } else if (accountID !== 0) {
    V1APIClient.fetchAccount(accountID)
      .then((resp: AxiosResponse<Account>) => {
        account = resp.data;
        originalAccount = { ...account };
      })
      .catch((err: AxiosError<ErrorResponse>) => {
        accountFetchError = err.message;
      });
  }
}
</script>

<div class="flex flex-wrap">
  <div class="w-full px-4">
    <div class="relative flex flex-col min-w-0 break-words w-full mb-6 shadow-lg rounded-lg bg-gray-200 border-0">
      <div class="rounded-t bg-white mb-0 px-6 py-6">
        <div class="text-center flex justify-between">
          <h6 class="text-gray-800 text-xl font-bold">
            {pageTranslations.title}
          </h6>
        </div>
      </div>

      <div class="flex-auto px-4 lg:px-10 py-10 pt-0">
        <div class="text-center flex justify-between">
          <h6 class="text-gray-500 text-sm mt-3 mb-6 font-bold uppercase">
            {pageTranslations.sectionLabels.info}
          </h6>
        </div>
        <div class="flex flex-wrap">
          <div class="w-full lg:w-6/12 px-4">
            <div class="relative w-full mb-3">
              <label class="block uppercase text-gray-700 text-xs font-bold mb-2" for="grid-username">
                {pageTranslations.inputLabels.name}
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
        {#if currentAuthStatus.adminPermissions !== null}
          <hr class="mt-6 border-b-1 border-gray-400" />

          <div class="text-center flex justify-between">
            <h6 class="text-gray-500 text-sm mt-3 mb-6 font-bold uppercase">
              {pageTranslations.sectionLabels.members}
            </h6>
            <button
              on:click="{console.log}"
              class="{false ? 'bg-blue-500' : 'bg-gray-300'} text-white active:bg-blue-600 font-bold uppercase text-xs rounded p-3 m-2"
              type="button"
            >
              {pageTranslations.buttons.saveMembers}
            </button>
          </div>

          <div class="flex flex-wrap">
            <div class="w-full lg:w-3/12 px-4">



              <table class="items-center w-full bg-transparent border-collapse">
                <thead>
                <tr>
                  {#each AccountUserMembership.headers(modelTranslations) as header}
                      <th class="px-6 align-middle border border-solid py-3 text-xs uppercase border-l-0 border-r-0 whitespace-no-wrap font-semibold text-left bg-gray-100 text-gray-600 border-gray-200">
                        {header.content}
                      </th>
                  {/each}
                </tr>
                </thead>
                <tbody>
                {#each account.members as membership}
                  <tr>
                    {#each AccountUserMembership.asRow(membership) as cell}
                        <td class="border-t-0 px-6 align-middle border-l-0 border-r-0 text-xs whitespace-no-wrap p-4" >
                          {cell.content}
                        </td>
                    {/each}
                  </tr>
                {/each}
                </tbody>
              </table>

            </div>
          </div>
          {/if}
      </div>
    </div>
  </div>
</div>
