<script lang="typescript">
import { navigate } from 'svelte-routing';
import * as faker from 'faker';
import format from 'string-format';
import Select from 'svelte-select';
import type { AxiosError, AxiosResponse } from 'axios';

import { UserSiteSettings, Webhook, WebhookCreationInput } from '../../types';
import { Logger } from '../../logger';
import { V1APIClient } from '../../apiClient';
import { contentTypes, frontendRoutes, methods } from '../../constants';
import { Superstore } from '../../stores';

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


let logger = new Logger().withDebugValue(
  'source',
  'src/components/creators/webhook.svelte',
);

function buildSelectOptionsFromStrings(...input: string[]): SelectOption[] {
  return input.map((x) => { return { value: x, label: x, } as SelectOption})
}

// local state
let webhook: WebhookCreationInput = new WebhookCreationInput();
let apiError: string = '';

// set up translations
let adminMode: boolean = false;
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
});

// Content-type
const validContentTypes: SelectOption[] = [
  {value: contentTypes.JSON, label: contentTypes.JSON },
  {value: contentTypes.XML, label: contentTypes.XML },
];

function selectContentType(value: SelectedValue) {
  webhook.contentType = value.detail.value;
  evaluateInputValidity();
}

function clearContentType() {
  webhook.contentType = '';
  evaluateInputValidity();
}

// Methods
const validMethods: SelectOption[] = buildSelectOptionsFromStrings(methods.PATCH, methods.PUT, methods.POST, methods.DELETE);

function selectMethod(value: SelectedValue) {
  webhook.method = value.detail.value;
  evaluateInputValidity();
}

function clearMethod() {
  webhook.method = '';
  evaluateInputValidity();
}

// Events
const validEvents: SelectOption[] = buildSelectOptionsFromStrings(...translationsToUse.validInputs.events);

function selectEvent(value: SelectedValues) {
  webhook.events = value.detail.map((value) => { return value.label });
  evaluateInputValidity();
}

function clearEvents() {
  webhook.events = [];
  evaluateInputValidity();
}

// Data Types
const validDataTypes: SelectOption[] = buildSelectOptionsFromStrings(...translationsToUse.validInputs.types)

function selectDataType(value: SelectedValues) {
  webhook.dataTypes = value.detail.map((value) => { return value.label });
  evaluateInputValidity();
}

function clearDataTypes() {
  webhook.dataTypes = [];
  evaluateInputValidity();
}


// Data Types
const validTopics: SelectOption[] = buildSelectOptionsFromStrings(...translationsToUse.validInputs.topics)

function selectTopic(value: SelectedValues) {
  webhook.topics = value.detail.map((value) => { return value.label });
  evaluateInputValidity();
}

function clearTopics() {
  webhook.topics = [];
  evaluateInputValidity();
}

// URL
let urlIsValid: boolean = false;
function validateURL(): void {
  urlIsValid = false;
  try {
    // only allow secure (lol) protocols
    urlIsValid = new URL(webhook.url).protocol.toLowerCase() === 'https:';
  } catch {
    urlIsValid = false;
  }

  evaluateInputValidity();
}

let webhookInputIsValid: boolean = false;
function evaluateInputValidity(): void {
  webhookInputIsValid = urlIsValid &&
                        webhook.name != "" &&
                        webhook.method != "" &&
                        webhook.contentType != "" &&
                        webhook.events.length !== 0 &&
                        webhook.dataTypes.length !== 0 &&
                        webhook.topics.length !== 0;
}

function createWebhook(): void {
  logger.debug(`createWebhook called`);
  evaluateInputValidity();

  if (superstore.frontendOnlyMode) {
    navigate(
      format(frontendRoutes.INDIVIDUAL_WEBHOOK, faker.random.number().toString()),
      {
        state: {},
        replace: true,
      },
    );
  } else {
    V1APIClient.createWebhook(webhook)
      .then((response: AxiosResponse<Webhook>) => {
        const newWebhook = response.data;
        logger
          .withValue('new_webhook_id', newWebhook.id)
          .debug(`navigating to webhook page via creation promise resolution`);
        navigate(
          format(frontendRoutes.INDIVIDUAL_WEBHOOK, newWebhook.id.toString()),
          {
            state: {},
            replace: true,
          },
        );
      })
      .catch((error: AxiosError) => {
        apiError = error.response?.data;
      });
  }
}
</script>

<div
  class="relative flex flex-col min-w-0 break-words bg-white w-full mb-6 shadow-lg rounded"
>
  <div class="rounded-t mb-0 px-4 py-3 bg-transparent justify-between ">
    <div class="flex flex-wrap webhooks-center">
      <div class="relative w-full max-w-full flex-grow flex-1">
        <h2 class="text-gray-800 text-xl font-semibold">
          {translationsToUse.model.actions.create}
        </h2>
      </div>
      <div class="flex w-full max-w-full flex-grow justify-end flex-1">
        <button
          class="{webhookInputIsValid ? 'bg-green-500 shadow-md' : 'bg-gray-200'} text-white border font-bold py-2 px-4 rounded"
          on:click="{createWebhook}"
        ><i class="fa fa-plus-circle"></i>
          Create</button>
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
          <Select items={validMethods} on:select={selectMethod} on:clear={clearMethod} />
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
          <Select items={validContentTypes} on:select={selectContentType} on:clear={clearContentType} />
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
          <Select items={validEvents} isMulti={true} on:select={selectEvent} on:clear={clearEvents} />
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
          <Select items={validDataTypes} isMulti={true} on:select={selectDataType} on:clear={clearDataTypes} />
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
          <Select items={validTopics} isMulti={true} on:select={selectTopic} on:clear={clearTopics} />
        </div>
      </div>
    </div>
  </div>
</div>
