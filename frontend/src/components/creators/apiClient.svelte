<script lang="typescript">
  import { navigate } from 'svelte-routing';
  import { onMount } from 'svelte';
  import type { AxiosError, AxiosResponse } from 'axios';

  import {
    APIClient,
    UserSiteSettings,
    UserStatus,
    AuditLogEntry,
    fakeAPIClientFactory,
    fakeAuditLogEntryFactory, APIClientCreationInput, Webhook,
  } from '../../types';
  import { Logger } from '../../logger';
  import { V1APIClient } from '../../apiClient';
  import AuditLogTable from '../core/auditLogTable/auditLogTable.svelte';
  import { frontendRoutes, statusCodes } from '../../constants';
  import { Superstore } from '../../stores';
  import format from "string-format";
  import * as faker from "faker";

  export let location: Location;

  // local state
  let input: APIClientCreationInput = new APIClientCreationInput();

  let logger = new Logger().withDebugValue(
    'source',
    'src/components/editors/apiClient.svelte',
  );

  let adminMode: boolean = false;
  let currentAuthStatus: UserStatus = new UserStatus();
  let currentSessionSettings = new UserSiteSettings();
  let apiError: string = '';
  let translationsToUse = currentSessionSettings.getTranslations().models.apiClient;

  let superstore = new Superstore({
    adminModeUpdateFunc: (value: boolean) => {
      adminMode = value;
    },
    userStatusStoreUpdateFunc: (value: UserStatus) => {
      currentAuthStatus = value;
    },
    sessionSettingsStoreUpdateFunc: (value: UserSiteSettings) => {
      currentSessionSettings = value;
      translationsToUse = currentSessionSettings.getTranslations().models.apiClient;
    },
  });

  function createAPIClient(): void {
    logger.debug(`createAPIClient called`);

    if (superstore.frontendOnlyMode) {
      navigate(
        format(frontendRoutes.INDIVIDUAL_WEBHOOK, faker.datatype.number().toString()),
        {
          state: {},
          replace: true,
        },
      );
    } else {
      V1APIClient.createAPIClient(input)
      .then((response: AxiosResponse<APIClient>) => {
        const newAPIClient = response.data;
        logger
        .withValue('new_api_client_id', newAPIClient.id)
        .debug(`navigating to webhook page via creation promise resolution`);
        navigate(
          format(frontendRoutes.INDIVIDUAL_API_CLIENT, newAPIClient.id.toString()),
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

<div>
  <div class="relative flex flex-col min-w-0 break-words bg-white w-full mb-6 shadow-lg rounded"
  >
    <div class="rounded-t mb-0 px-4 py-3 bg-transparent justify-between ">
      <div class="flex flex-wrap items-center">
        <div class="relative w-full max-w-full flex-grow flex-1">
          New API Client
        </div>

        <div class="flex w-full max-w-full flex-grow justify-end flex-1">
          <button class="{input.complete() ? 'bg-green-500 shadow-md' : 'bg-gray-200'} text-white border font-bold py-2 px-4 rounded"
                  on:click="{createAPIClient}"
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
                  for="grid-username"
          >
            {translationsToUse.labels.username}
          </label>
          <input
                  class="appearance-none block w-full text-gray-700 border border-red-500 rounded py-3 px-4 mb-3 leading-tight focus:outline-none focus:bg-white"
                  id="grid-username"
                  type="text"
                  bind:value="{input.username}"
          />
          <!--  <p class="text-red-500 text-xs italic">Please fill out this field.</p>-->
        </div>
        <div class="w-full md:w-1/2 px-3 mb-6 md:mb-0">
          <label
                  class="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2"
                  for="grid-password"
          >
            {translationsToUse.labels.password}
          </label>
          <input
                  class="appearance-none block w-full text-gray-700 border border-red-500 rounded py-3 px-4 mb-3 leading-tight focus:outline-none focus:bg-white"
                  id="grid-password"
                  type="password"
                  bind:value="{input.password}"
          />
          <!--  <p class="text-red-500 text-xs italic">Please fill out this field.</p>-->
        </div>
        <div class="w-full md:w-1/2 px-3 mb-6 md:mb-0">
          <label
                  class="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2"
                  for="grid-totp-token"
          >
            {translationsToUse.labels.totpToken}
          </label>
          <input
                  class="appearance-none block w-full text-gray-700 border border-red-500 rounded py-3 px-4 mb-3 leading-tight focus:outline-none focus:bg-white"
                  id="grid-totp-token"
                  type="text"
                  bind:value="{input.totpToken}"
          />
          <!--  <p class="text-red-500 text-xs italic">Please fill out this field.</p>-->
        </div>
        <div class="w-full md:w-1/2 px-3 mb-6 md:mb-0">
          <label
                  class="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2"
                  for="grid-client-name"
          >
            {translationsToUse.labels.name}
          </label>
          <input
                  class="appearance-none block w-full text-gray-700 border border-red-500 rounded py-3 px-4 mb-3 leading-tight focus:outline-none focus:bg-white"
                  id="grid-client-name"
                  type="text"
                  bind:value="{input.name}"
          />
          <!--  <p class="text-red-500 text-xs italic">Please fill out this field.</p>-->
        </div>
      </div>
    </div>
  </div>

  {#if currentAuthStatus.adminPermissions !== null && adminMode}
    <AuditLogTable entryFetchFunc="{fetchAuditLogEntries}" />
  {/if}
</div>
