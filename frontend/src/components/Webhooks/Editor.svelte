<script lang="typescript">
import { navigate } from 'svelte-routing';
import { onMount } from 'svelte';
import { AxiosError, AxiosResponse } from 'axios';

import AuditLogTable from '../AuditLogTable/AuditLogTable.svelte';

import {
  Webhook,
  UserSiteSettings,
  UserStatus,
  AuditLogEntry,
  fakeWebhookFactory,
  fakeAuditLogEntryFactory,
} from '../../types';
import { V1APIClient } from '../../apiClient';
import { frontendRoutes, statusCodes } from '../../constants';
import { Superstore } from '../../stores';

import { Logger } from '../../logger';
import { renderUnixTime } from '../../utils';

export let webhookID: number = 0;

// local state
let originalWebhook: Webhook = new Webhook();
let webhook: Webhook = new Webhook();
let webhookRetrievalError: string = '';
let auditLogEntriesRetrievalError: string = '';
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
let adminMode: boolean = false;
let currentUserStatus = new UserStatus();
let currentSessionSettings = new UserSiteSettings();
let translationsToUse = currentSessionSettings.getTranslations().models.webhook;
const superstore = new Superstore({
  adminModeUpdateFunc: (value: boolean) => {
    adminMode = value;
  },
  sessionSettingsStoreUpdateFunc: (value: UserSiteSettings) => {
    currentSessionSettings = value;
    translationsToUse = currentSessionSettings.getTranslations().models.webhook;
  },
  userStatusStoreUpdateFunc: (value: UserStatus) => {
    currentUserStatus = value;
  },
});

function fetchWebhook(): void {
  logger.debug(`fetchWebhook called`);

  if (webhookID === 0) {
    throw new Error('webhookID cannot be zero!');
  }

  if (superstore.frontendOnlyMode) {
    webhook = fakeWebhookFactory.build();
    originalWebhook = { ...webhook };
  } else {
    V1APIClient.fetchWebhook(webhookID)
      .then((response: AxiosResponse<Webhook>) => {
        webhook = { ...response.data };
        originalWebhook = { ...response.data };
      })
      .catch((error: AxiosError) => {
        webhookRetrievalError = error.response?.data;
      });
  }

  fetchAuditLogEntries();
}

function saveWebhook(): void {
  logger.debug(`saveWebhook called`);

  if (webhookID === 0) {
    throw new Error('webhookID cannot be zero!');
  } else if (!needsToBeSaved) {
    throw new Error('no changes to save!');
  }

  if (superstore.frontendOnlyMode) {
    originalWebhook = { ...webhook };
    needsToBeSaved = false;
  } else {
    V1APIClient.saveWebhook(webhook)
      .then((response: AxiosResponse<Webhook>) => {
        webhook = { ...response.data };
        originalWebhook = { ...response.data };
        needsToBeSaved = false;
        fetchAuditLogEntries();
      })
      .catch((error: AxiosError) => {
        webhookRetrievalError = error.response?.data;
      });
  }
}

function deleteWebhook(): void {
  logger.debug(`deleteWebhook called`);

  if (webhookID === 0) {
    throw new Error('webhookID cannot be zero!');
  }

  if (superstore.frontendOnlyMode) {
    navigate(frontendRoutes.USER_LIST_WEBHOOKS, { state: {}, replace: true });
  } else {
    V1APIClient.deleteWebhook(webhookID)
      .then((response: AxiosResponse<Webhook>) => {
        if (response.status === statusCodes.NO_CONTENT) {
          logger.debug(
            `navigating to ${frontendRoutes.USER_LIST_WEBHOOKS} because via deletion promise resolution`,
          );
          navigate(frontendRoutes.USER_LIST_WEBHOOKS, { state: {}, replace: true });
        }
      })
      .catch((error: AxiosError) => {
        webhookRetrievalError = error.response?.data;
      });
  }
}

function fetchAuditLogEntries(): void {
  logger.debug(`fetchAuditLogEntries called`);

  if (webhookID === 0) {
    throw new Error('webhookID cannot be zero!');
  }

  if (!adminMode) {
    return;
  }

  if (superstore.frontendOnlyMode) {
    auditLogEntries = fakeAuditLogEntryFactory.buildList(10);
  } else {
    V1APIClient.fetchAuditLogEntriesForWebhook(webhookID)
      .then((response: AxiosResponse<AuditLogEntry[]>) => {
        auditLogEntries = response.data;
        logger.withValue('entries', auditLogEntries).debug('entries fetched');
      })
      .catch((error: AxiosError) => {
        auditLogEntriesRetrievalError = error.response?.data;
      });
  }
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
            for="grid-name"
          >
            {translationsToUse.labels.name}
          </label>
          <input
            class="appearance-none block w-full text-gray-700 border border-red-500 rounded py-3 px-4 mb-3 leading-tight focus:outline-none focus:bg-white"
            id="grid-name"
            type="text"
            on:keyup="{evaluateChanges}"
            bind:value="{webhook.name}"
          />
          <label
            class="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2"
            for="grid-content-type"
          >
            {translationsToUse.labels.contentType}
          </label>
          <input
            class="appearance-none block w-full text-gray-700 border border-red-500 rounded py-3 px-4 mb-3 leading-tight focus:outline-none focus:bg-white"
            id="grid-content-type"
            type="text"
            on:keyup="{evaluateChanges}"
            bind:value="{webhook.contentType}"
          />

          <label
            class="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2"
            for="grid-url"
          >
            {translationsToUse.labels.url}
          </label>
          <input
            class="appearance-none block w-full text-gray-700 border border-red-500 rounded py-3 px-4 mb-3 leading-tight focus:outline-none focus:bg-white"
            id="grid-url"
            type="text"
            on:keyup="{evaluateChanges}"
            bind:value="{webhook.url}"
          />

          <label
            class="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2"
            for="grid-method"
          >
            {translationsToUse.labels.method}
          </label>
          <input
            class="appearance-none block w-full text-gray-700 border border-red-500 rounded py-3 px-4 mb-3 leading-tight focus:outline-none focus:bg-white"
            id="grid-method"
            type="text"
            on:keyup="{evaluateChanges}"
            bind:value="{webhook.method}"
          />

          <label
            class="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2"
            for="grid-events"
          >
            {translationsToUse.labels.events}
          </label>
          <ul class="m-2" id="grid-events">
            {#each webhook.events as event}
              <li>{event}</li>
            {/each}
          </ul>

          <label
            class="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2"
            for="grid-data-types"
          >
            {translationsToUse.labels.dataTypes}
          </label>
          <ul class="m-2" id="grid-data-types">
            {#each webhook.dataTypes as dataType}
              <li>{dataType}</li>
            {/each}
          </ul>

          <label
            class="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2"
            for="grid-topics"
          >
            {translationsToUse.labels.topics}
          </label>
          <ul class="m-2" id="grid-topics">
            {#each webhook.topics as topic}
              <li>{topic}</li>
            {/each}
          </ul>

          <label
            class="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2"
            for="grid-created-on"
          >
            {translationsToUse.labels.createdOn}
          </label>
          <p id="grid-created-on">{renderUnixTime(webhook.createdOn)}</p>
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
    <AuditLogTable entryFetchFunc="{fetchAuditLogEntries}" />
  {/if}
</div>
