<script lang="typescript">
import { navigate } from 'svelte-routing';
import * as faker from 'faker';
import format from 'string-format';
import Select from 'svelte-select';
import { isWebUri } from 'valid-url';
import { AxiosError, AxiosResponse } from 'axios';

import { Webhook, WebhookCreationInput, UserSiteSettings } from '../../types';
import { Logger } from '../../logger';
import { V1APIClient } from '../../apiClient';
import {frontendRoutes, methods} from '../../constants';
import { Superstore } from '../../stores';

// local state
let webhook: WebhookCreationInput = new WebhookCreationInput();
let apiError: string = '';

let logger = new Logger().withDebugValue(
  'source',
  'src/components/Creators/Webhook.svelte',
);

interface selectOptions {
  value: string;
  label: string;
}

// set up translations
let currentSessionSettings = new UserSiteSettings();
let translationsToUse = currentSessionSettings.getTranslations().models.webhook;
const superstore = new Superstore({
  sessionSettingsStoreUpdateFunc: (value: UserSiteSettings) => {
    currentSessionSettings = value;
    translationsToUse = currentSessionSettings.getTranslations().models.webhook;
  },
});

let selectedContentType: string = '';
const validContentTypes: selectOptions[] = [
  {value: "application/json", label: "application/json" },
  {value: "application/xml", label: "application/xml" },
];


let selectedMethod: string = '';
const validMethods: selectOptions[] = [
  {value: methods.PATCH, label: methods.PATCH },
  {value: methods.PUT, label: methods.PUT },
  {value: methods.POST, label: methods.POST },
  {value: methods.DELETE, label: methods.DELETE },
];

let selectedEvents: any[] = new Array();
const validEvents: selectOptions[] = [
  {value: "Create", label: "Create" },
  {value: "Update", label: "Update" },
  {value: "Delete", label: "Delete" },
];

let selectedDataTypes: string[] = [];
const validDataTypes: selectOptions[]  = [
  {value: "Item", label: "Item" },
];


let urlIsValid: boolean = false;
function validateURL(): void {
  logger.debug("validateURL called")
  urlIsValid = isWebUri(webhook.url) !== '';
  evaluateInputValidity();
}

let canProceed: boolean = false;
function evaluateInputValidity(): void {
  canProceed = urlIsValid && true;
}

function createWebhook(): void {
  logger.debug(`createWebhook called`);
  evaluateInputValidity();

  webhook.contentType = selectedContentType;
  webhook.method = selectedMethod;

  console.dir(webhook);
  console.dir(selectedEvents);

  if (!canProceed) {
    return;
  }

  if (superstore.frontendOnlyMode) {
    navigate(
      format(frontendRoutes.INDIVIDUAL_ITEM, faker.random.number().toString()),
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
          format(frontendRoutes.INDIVIDUAL_ITEM, newWebhook.id.toString()),
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
          {translationsToUse.actions.create}
        </h2>
      </div>
      <div class="flex w-full max-w-full flex-grow justify-end flex-1">
        <button
          class="bg-green-500 hover:bg-green-700 text-white font-bold py-2 px-4 rounded"
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
          {translationsToUse.labels.name}
        </label>
        <input
          class="appearance-none block w-full text-gray-700 border border-red-500 rounded py-3 px-4 mb-3 leading-tight focus:outline-none focus:bg-white"
          id="grid-first-name"
          type="text"
          bind:value="{webhook.name}"
        />
      </div>
      <div class="w-full md:w-1/2 px-3 mb-6 md:mb-0">
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
          on:blur={validateURL}
          bind:value="{webhook.url}"
        />
      </div>
      <div class="w-full md:w-1/2 px-3 mb-6 md:mb-0">
        <label
          class="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2"
          for="grid-method"
        >
          {translationsToUse.labels.method}
        </label>

        <div  id="grid-method">
          <Select items={validMethods} bind:selectedMethod />
        </div>
      </div>
      <div class="w-full md:w-1/2 px-3 mb-6 md:mb-0">
        <label
          class="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2"
          for="grid-content-type"
        >
          {translationsToUse.labels.contentType}
        </label>

        <div  id="grid-content-type">
          <Select items={validContentTypes} bind:selectedContentType />
        </div>
      </div>
      <div class="w-full md:w-1/2 px-3 mb-6 md:mb-0">
        <label
          class="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2"
          for="grid-events"
        >
          {translationsToUse.labels.events}
        </label>

        <div  id="grid-events">
          <Select items={validEvents} isMulti={true} bind:selectedEvents />
        </div>
      </div>
      <div class="w-full md:w-1/2 px-3 mb-6 md:mb-0">
        <label
          class="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2"
          for="grid-data-types"
        >
          {translationsToUse.labels.dataTypes}
        </label>

        <div  id="grid-data-types">
          <Select items={validDataTypes} isMulti={true} bind:selectedDataTypes />
        </div>
      </div>
      <div class="w-full md:w-1/2 px-3 mb-6 md:mb-0">
        <label
          class="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2"
          for="grid-topics"
        >
          {translationsToUse.labels.topics}
        </label>
        <input
          class="appearance-none block w-full text-gray-700 border border-red-500 rounded py-3 px-4 mb-3 leading-tight focus:outline-none focus:bg-white"
          id="grid-topics"
          type="text"
          bind:value="{webhook.topics}"
        />
      </div>
    </div>
  </div>
</div>
