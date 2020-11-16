<script lang="typescript">
import { navigate } from 'svelte-routing';
import { onDestroy, onMount } from 'svelte';
import { AxiosError, AxiosResponse } from 'axios';

import {
  Webhook,
  UserSiteSettings,
  UserStatus,
  AuditLogEntry,
} from '../../types';
import { Logger } from '../../logger';
import { V1APIClient } from '../../apiClient';
import { translations } from '../../i18n';
import { sessionSettingsStore, userStatusStore } from '../../stores';
import AuditLogTable from '../AuditLogTable/AuditLogTable.svelte';
import { frontendRoutes, statusCodes } from '../../constants';

export let webhookID: number = 0;

// local state
let originalWebhook: Webhook = new Webhook();
let webhook: Webhook = new Webhook();
let webhookRetrievalError: string = '';
let needsToBeSaved: boolean = false;
let auditLogEntries: AuditLogEntry[] = [];

function evaluateChanges() {
  needsToBeSaved = !Webhook.areEqual(webhook, originalWebhook);
}

onMount(fetchWebhook);

let logger = new Logger().withDebugValue(
  'source',
  'src/components/Editors/Webhook.svelte',
);

// set up translations
let currentSessionSettings = new UserSiteSettings();
let translationsToUse = translations.messagesFor(
  currentSessionSettings.language,
).models.webhook;
const unsubscribeFromSettingsUpdates = sessionSettingsStore.subscribe(
  (value: UserSiteSettings) => {
    currentSessionSettings = value;
    translationsToUse = translations.messagesFor(
      currentSessionSettings.language,
    ).models.webhook;
  },
);
// onDestroy(unsubscribeFromSettingsUpdates);

// set up user status sync
let currentUserStatus = new UserStatus();
const unsubscribeFromUserStatusUpdates = userStatusStore.subscribe(
  (value: UserStatus) => {
    currentUserStatus = value;
  },
);
// onDestroy(unsubscribeFromUserStatusUpdates);

function fetchWebhook(): void {
  logger.debug(`fetchWebhook called`);

  if (webhookID === 0) {
    throw new Error('webhookID cannot be zero!');
  }

  V1APIClient.fetchWebhook(webhookID)
    .then((response: AxiosResponse<Webhook>) => {
      webhook = { ...response.data };
      originalWebhook = { ...response.data };
    })
    .catch((error: AxiosError) => {
      if (error.response) {
        if (error.response.data) {
          webhookRetrievalError = error.response.data;
        }
      }
    });

  fetchAuditLogEntries();
}

function saveWebhook(): void {
  logger.debug(`saveWebhook called`);

  if (webhookID === 0) {
    throw new Error('webhookID cannot be zero!');
  } else if (!needsToBeSaved) {
    throw new Error('no changes to save!');
  }

  V1APIClient.saveWebhook(webhook)
    .then((response: AxiosResponse<Webhook>) => {
      webhook = { ...response.data };
      originalWebhook = { ...response.data };
      needsToBeSaved = false;
      fetchAuditLogEntries();
    })
    .catch((error: AxiosError) => {
      if (error.response && error.response.data) {
        webhookRetrievalError = error.response.data;
      }
    });
}

function deleteWebhook(): void {
  logger.debug(`deleteWebhook called`);

  if (webhookID === 0) {
    throw new Error('webhookID cannot be zero!');
  }

  V1APIClient.deleteWebhook(webhookID)
    .then((response: AxiosResponse<Webhook>) => {
      if (response.status === statusCodes.NO_CONTENT) {
        logger.debug(
          `navigating to ${frontendRoutes.LIST_WEBHOOKS} because via deletion promise resolution`,
        );
        navigate(frontendRoutes.LIST_WEBHOOKS, { state: {}, replace: true });
      }
    })
    .catch((error: AxiosError) => {
      if (error.response) {
        if (error.response.data) {
          webhookRetrievalError = error.response.data;
        }
      }
    });
}

function fetchAuditLogEntries(): void {
  logger.debug(`deleteWebhook called`);

  if (webhookID === 0) {
    throw new Error('webhookID cannot be zero!');
  }

  V1APIClient.fetchAuditLogEntriesForWebhook(webhookID)
    .then((response: AxiosResponse<AuditLogEntry[]>) => {
      auditLogEntries = response.data;
      logger.withValue('entries', auditLogEntries).debug('entries fetched');
    })
    .catch((error: AxiosError) => {
      if (error.response) {
        if (error.response.data) {
          webhookRetrievalError = error.response.data;
        }
      }
    });
}
</script>

<div>
  <div
    class="relative flex flex-col min-w-0 break-words bg-white w-full mb-6 shadow-lg rounded"
  >
    <div class="rounded-t mb-0 px-4 py-3 bg-transparent justify-between ">
      <div class="flex flex-wrap webhooks-center">
        <div class="relative w-full max-w-full flex-grow flex-1">
          {#if originalWebhook.id !== 0}
            <h2 class="text-gray-800 text-xl font-semibold">
              #{originalWebhook.id}:
              {originalWebhook.name}
            </h2>
          {/if}
        </div>
        <div class="flex w-full max-w-full flex-grow justify-end flex-1">
          <button
            class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded {needsToBeSaved ? '' : 'opacity-50 cursor-not-allowed'}"
            on:click="{saveWebhook}"
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
            bind:value="{webhook.name}"
          />
          <!--  <p class="text-red-500 text-xs italic">Please fill out this field.</p>-->
        </div>
        <div class="w-full md:w-1/2 px-3">
          <label
            class="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2"
            for="grid-last-name"
          >
            {translationsToUse.labels.details}
          </label>
          <input
            class="appearance-none block w-full text-gray-700 border border-gray-200 rounded py-3 px-4 leading-tight focus:outline-none focus:bg-white focus:border-gray-500"
            id="grid-last-name"
            type="text"
            on:keyup="{evaluateChanges}"
            bind:value="{webhook.details}"
          />
        </div>
        <div
          class="flex w-full mr-3 mt-4 max-w-full flex-grow justify-end flex-1"
        >
          <button
            class="bg-red-500 hover:bg-red-700 text-white font-bold py-2 px-4 rounded"
            on:click="{deleteWebhook}"
          ><i class="fa fa-trash-alt"></i>
            Delete</button>
        </div>
      </div>
    </div>
  </div>

  {#if currentUserStatus.isAdmin}
    <AuditLogTable entries="{auditLogEntries}" />
  {/if}
</div>
