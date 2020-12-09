<script lang="typescript">
import { navigate } from 'svelte-routing';
import Select from 'svelte-select';
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
import { contentTypes, frontendRoutes, methods, statusCodes } from '../../constants';
import { Superstore } from '../../stores';
import { Logger } from '../../logger';

declare interface SelectOption {
  value: string;
  label: string;
}

declare interface SelectedValue {
  detail: SelectOption
}

declare interface SelectedValues {
  detail: SelectOption[]
}

export let webhookID: number = 0;

// local state
let originalWebhook: Webhook = new Webhook();
let webhook: Webhook = new Webhook();
let webhookRetrievalError: string = '';
let auditLogEntriesRetrievalError: string = '';
let needsToBeSaved: boolean = false;
let auditLogEntries: AuditLogEntry[] = [];

function evaluateChanges() {
  evaluateInputValidity()
  needsToBeSaved = !Webhook.areEqual(webhook, originalWebhook) && webhookIsValid;
}

function buildSelectOptionsFromStrings(...input: string[]): SelectOption[] {
  return input.map((x) => { return { value: x, label: x, } as SelectOption})
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
let translationsToUse = currentSessionSettings.getTranslations().pages.webhookCreationPage;
const superstore = new Superstore({
  adminModeUpdateFunc: (value: boolean) => {
    adminMode = value;
  },
  sessionSettingsStoreUpdateFunc: (value: UserSiteSettings) => {
    currentSessionSettings = value;
    translationsToUse = currentSessionSettings.getTranslations().pages.webhookCreationPage;
  },
  userStatusStoreUpdateFunc: (value: UserStatus) => {
    currentUserStatus = value;
  },
});

let webhookIsValid: boolean = false;
function evaluateInputValidity(): void {
  webhookIsValid = urlIsValid &&
    webhook.name != "" &&
    webhook.method != "" &&
    webhook.contentType != "" &&
    webhook.events.length !== 0 &&
    webhook.dataTypes.length !== 0;
}

// Content-type
const validContentTypes: SelectOption[] = [
  {value: contentTypes.JSON, label: contentTypes.JSON },
  {value: contentTypes.XML, label: contentTypes.XML },
];

function selectContentType(value: SelectedValue) {
  webhook.contentType = value.detail.value;
  evaluateChanges();
}

function clearContentType() {
  webhook.contentType = '';
  evaluateChanges();
}

// Methods
const validMethods: SelectOption[] = buildSelectOptionsFromStrings(methods.PATCH, methods.PUT, methods.POST, methods.DELETE);

function selectMethod(value: SelectedValue) {
  webhook.method = value.detail.value;
  evaluateChanges();
}

function clearMethod() {
  webhook.method = '';
  evaluateChanges();
}

// Events
const validEvents: SelectOption[] = buildSelectOptionsFromStrings(...translationsToUse.validInputs.events);

function selectEvent(value: SelectedValues) {
  webhook.events = value.detail.map((value) => { return value.label });
  evaluateChanges();
}

function clearEvents() {
  webhook.events = [];
  evaluateChanges();
}

// Data Types
const validDataTypes: SelectOption[] = buildSelectOptionsFromStrings(...translationsToUse.validInputs.types)

function selectDataType(value: SelectedValues) {
  webhook.dataTypes = value.detail.map((value) => { return value.label });
  evaluateChanges();
}

function clearDataTypes() {
  webhook.dataTypes = [];
  evaluateChanges();
}

// Topics
const validTopics: SelectOption[] = buildSelectOptionsFromStrings(...translationsToUse.validInputs.types)

function selectTopic(value: SelectedValues) {
  webhook.topics = value.detail.map((value) => { return value.label });
  evaluateChanges();
}

function clearTopics() {
  webhook.topics = [];
  evaluateChanges();
}

// URL
let urlIsValid: boolean = false;
function validateURL(): void {
  logger.debug("validateURL called")

  urlIsValid = false;
  try {
    // only allow secure (lol) protocols
    urlIsValid = new URL(webhook.url).protocol.toLowerCase() === 'https:';
  } catch {
    urlIsValid = false;
  }

  evaluateChanges();
}

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
            for="grid-first-name"
          >
            {translationsToUse.model.labels.name}
          </label>
          <input
            class="appearance-none block w-full text-gray-700 border border-red-500 rounded py-3 px-4 mb-3 leading-tight focus:outline-none focus:bg-white"
            id="grid-first-name"
            type="text"
            on:blur={evaluateInputValidity}
            bind:value="{webhook.name}"
          />
        </div>
        <div class="w-full md:w-1/2 px-3 mb-6 md:mb-0">
          <label
            class="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2"
            for="grid-url"
          >
            {translationsToUse.model.labels.url}
          </label>
          <input
            class="appearance-none block w-full text-gray-700 border border-red-500 rounded py-3 px-4 mb-3 leading-tight focus:outline-none focus:bg-white"
            id="grid-url"
            type="text"
            on:blur={validateURL}
            bind:value="{webhook.url}"
          />
        </div>
        <div class="w-full md:w-1/2 px-3 mb-6 md:mb-0">
          <label
            class="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2"
            for="grid-method"
          >
            {translationsToUse.model.labels.method}
          </label>

          <div  id="grid-method">
            <Select items={validMethods} selectedValue={webhook.method} on:select={selectMethod} on:clear={clearMethod} />
          </div>
        </div>
        <div class="w-full md:w-1/2 px-3 mb-6 md:mb-0">
          <label
            class="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2"
            for="grid-content-type"
          >
            {translationsToUse.model.labels.contentType}
          </label>

          <div  id="grid-content-type">
            <Select items={validContentTypes} selectedValue={webhook.contentType} on:select={selectContentType} on:clear={clearContentType} />
          </div>
        </div>
        <div class="w-full md:w-1/2 px-3 mb-6 md:mb-0">
          <label
            class="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2"
            for="grid-events"
          >
            {translationsToUse.model.labels.events}
          </label>

          <div  id="grid-events">
            <Select items={validEvents} selectedValue={webhook.events} isMulti={true} on:select={selectEvent} on:clear={clearEvents} />
          </div>
        </div>
        <div class="w-full md:w-1/2 px-3 mb-6 md:mb-0">
          <label
            class="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2"
            for="grid-data-types"
          >
            {translationsToUse.model.labels.dataTypes}
          </label>

          <div  id="grid-data-types">
            <Select items={validDataTypes} selectedValue={webhook.dataTypes} isMulti={true} on:select={selectDataType} on:clear={clearDataTypes} />
          </div>
        </div>
        <div class="w-full md:w-1/2 px-3 mb-6 md:mb-0">
          <label
            class="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2"
            for="grid-topics"
          >
            {translationsToUse.model.labels.topics}
          </label>

          <div  id="grid-topics">
            <Select disabled={true} items={validDataTypes} selectedValue={webhook.topics} isMulti={true} on:select={selectTopic} on:clear={clearTopics} />
          </div>
        </div>
      </div>
    </div>
  </div>

  {#if currentUserStatus.isAdmin}
    <AuditLogTable entryFetchFunc="{fetchAuditLogEntries}" />
  {/if}
</div>
