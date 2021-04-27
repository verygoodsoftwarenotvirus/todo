<script lang="typescript">
import { navigate } from 'svelte-routing';
import { onMount } from 'svelte';
import type { AxiosError, AxiosResponse } from 'axios';

import {
  Account,
  UserSiteSettings,
  UserStatus,
  AuditLogEntry,
  fakeAccountFactory,
  fakeAuditLogEntryFactory,
} from '../../types';
import { Logger } from '../../logger';
import { V1APIClient } from '../../apiClient';
import AuditLogTable from '../core/auditLogTable/auditLogTable.svelte';
import { frontendRoutes, statusCodes } from '../../constants';
import { Superstore } from '../../stores';

export let accountID: number = 0;

// local state
let originalAccount: Account = new Account();
let account: Account = new Account();
let accountRetrievalError: string = '';
let auditLogEntriesRetrievalError: string = '';
let needsToBeSaved: boolean = false;
let auditLogEntries: AuditLogEntry[] = [];

function evaluateChanges() {
  needsToBeSaved = !Account.areEqual(account, originalAccount);
}

onMount(fetchAccount);

let logger = new Logger().withDebugValue(
  'source',
  'src/components/editors/account.svelte',
);

let adminMode: boolean = false;
let currentAuthStatus: UserStatus = new UserStatus();
let currentSessionSettings = new UserSiteSettings();
let translationsToUse = currentSessionSettings.getTranslations().models.account;

let superstore = new Superstore({
  adminModeUpdateFunc: (value: boolean) => {
    adminMode = value;
  },
  userStatusStoreUpdateFunc: (value: UserStatus) => {
    currentAuthStatus = value;
  },
  sessionSettingsStoreUpdateFunc: (value: UserSiteSettings) => {
    currentSessionSettings = value;
    translationsToUse = currentSessionSettings.getTranslations().models.account;
  },
});

function fetchAccount(): void {
  logger.debug(`fetchAccount called`);

  if (accountID === 0) {
    throw new Error('id cannot be zero!');
  }

  if (superstore.frontendOnlyMode) {
    const i = fakeAccountFactory.build();
    account = { ...i };
    originalAccount = { ...i };
  } else if (accountID !== 0){
    V1APIClient.fetchAccount(accountID)
      .then((response: AxiosResponse<Account>) => {
        account = { ...response.data };
        originalAccount = { ...response.data };
      })
      .catch((error: AxiosError) => {
        accountRetrievalError = error.response?.data;
      });
  }

  fetchAuditLogEntries();
}

function saveAccount(): void {
  logger.debug(`saveAccount called`);

  if (accountID === 0) {
    throw new Error('id cannot be zero!');
  } else if (!needsToBeSaved) {
    throw new Error('no changes to save!');
  }

  if (superstore.frontendOnlyMode) {
    needsToBeSaved = false;
    originalAccount = { ...account };
  } else {
    V1APIClient.saveAccount(account)
      .then((response: AxiosResponse<Account>) => {
        account = { ...response.data };
        originalAccount = { ...response.data };
        needsToBeSaved = false;
        fetchAuditLogEntries();
      })
      .catch((error: AxiosError) => {
        accountRetrievalError = error.response?.data;
      });
  }
}

function deleteAccount(): void {
  logger.debug(`deleteAccount called`);

  if (accountID === 0) {
    throw new Error('id cannot be zero!');
  }

  if (superstore.frontendOnlyMode) {
    navigate(frontendRoutes.LIST_ITEMS, { state: {}, replace: true });
  } else {
    V1APIClient.deleteAccount(accountID)
      .then((response: AxiosResponse<Account>) => {
        if (response?.status === statusCodes.NO_CONTENT) {
          logger.debug(
            `navigating to ${frontendRoutes.LIST_ITEMS} via deletion promise resolution`,
          );
          navigate(frontendRoutes.LIST_ITEMS, { state: {}, replace: true });
        }
      })
      .catch((error: AxiosError) => {
        auditLogEntriesRetrievalError = error.response?.data;
      });
  }
}

function fetchAuditLogEntries(): Promise<AxiosResponse<AuditLogEntry[]>> {
  logger.debug(`fetchAuditLogEntries called`);

  if (accountID === 0) {
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
    return V1APIClient.fetchAuditLogEntriesForAccount(accountID);
  }
}
</script>

<div>
  <div
    class="relative flex flex-col min-w-0 break-words bg-white w-full mb-6 shadow-lg rounded"
  >
    <div class="rounded-t mb-0 px-4 py-3 bg-transparent justify-between ">
      <div class="flex flex-wrap accounts-center">
        <div class="relative w-full max-w-full flex-grow flex-1">
          {#if originalAccount.id !== 0}
            <h2 class="text-gray-800 text-xl font-semibold">
              #{originalAccount.id}:
              {originalAccount.name}
            </h2>
          {/if}
        </div>
        <div class="flex w-full max-w-full flex-grow justify-end flex-1">
          <button
            class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded {needsToBeSaved ? '' : 'opacity-50 cursor-not-allowed'}"
            on:click="{saveAccount}"
          ><i class="fa fa-save"></i>
            Save</button>
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
            {translationsToUse.labels.name}
          </label>
          <input
            class="appearance-none block w-full text-gray-700 border border-red-500 rounded py-3 px-4 mb-3 leading-tight focus:outline-none focus:bg-white"
            id="grid-first-name"
            type="text"
            on:keyup="{evaluateChanges}"
            bind:value="{account.name}"
          />
          <!--  <p class="text-red-500 text-xs italic">Please fill out this field.</p>-->
        </div>
        <div class="w-full md:w-1/2 px-3">
          <label
            class="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2"
            for="grid-last-name"
          >
            {translationsToUse.labels.accountSubscriptionPlanID}
          </label>
          <input
            class="appearance-none block w-full text-gray-700 border border-gray-200 rounded py-3 px-4 leading-tight focus:outline-none focus:bg-white focus:border-gray-500"
            id="grid-last-name"
            type="text"
            on:keyup="{evaluateChanges}"
            bind:value="{account.accountSubscriptionPlanID}"
          />
        </div>
        <div
          class="flex w-full mr-3 mt-4 max-w-full flex-grow justify-end flex-1"
        >
          <button
            class="bg-red-500 hover:bg-red-700 text-white font-bold py-2 px-4 rounded"
            on:click="{deleteAccount}"
          ><i class="fa fa-trash-alt"></i>
            Delete</button>
        </div>
      </div>
    </div>
  </div>

  {#if currentAuthStatus.adminPermissions !== null && adminMode}
    <AuditLogTable entryFetchFunc="{fetchAuditLogEntries}" />
  {/if}
</div>
