<script lang="typescript">
  import { navigate } from 'svelte-routing';
  import { onDestroy, onMount } from 'svelte';
  import { AxiosError, AxiosResponse } from 'axios';

  import {
    OAuth2Client,
    UserSiteSettings,
    UserStatus,
    AuditLogEntry,
  } from '../../../types';
  import { Logger } from '../../../logger';
  import { V1APIClient } from '../../../requests';
  import { translations } from '../../../i18n';
  import { sessionSettingsStore, userStatusStore } from '../../../stores';
  import AuditLogTable from '../../AuditLogTable/AuditLogTable.svelte';

  export let id: number = 0;

  // local state
  let originalOAuth2Client: OAuth2Client = new OAuth2Client();
  let oauth2Client: OAuth2Client = new OAuth2Client();
  let oauth2ClientRetrievalError: string = '';
  let needsToBeSaved: boolean = false;
  let auditLogEntries: AuditLogEntry[] = [];

  function evaluateChanges() {
    needsToBeSaved = !OAuth2Client.areEqual(oauth2Client, originalOAuth2Client);
  }

  onMount(fetchOAuth2Client);

  let logger = new Logger().withDebugValue(
    'source',
    'src/components/Types/OAuth2Clients/Editor.svelte',
  );

  // set up translations
  let currentSessionSettings = new UserSiteSettings();
  let translationsToUse = translations.messagesFor(
    currentSessionSettings.language,
  ).models.oauth2Client;
  const unsubscribeFromSettingsUpdates = sessionSettingsStore.subscribe(
    (value: UserSiteSettings) => {
      currentSessionSettings = value;
      translationsToUse = translations.messagesFor(
        currentSessionSettings.language,
      ).models.oauth2Client;
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

  function fetchOAuth2Client(): void {
    logger.debug(`fetchOAuth2Client called`);

    if (id === 0) {
      throw new Error('id cannot be zero!');
    }

    V1APIClient.fetchOAuth2Client(id)
      .then((response: AxiosResponse<OAuth2Client>) => {
        oauth2Client = { ...response.data };
        originalOAuth2Client = { ...response.data };
      })
      .catch((error: AxiosError) => {
        if (error.response) {
          if (error.response.data) {
            oauth2ClientRetrievalError = error.response.data;
          }
        }
      });

    fetchAuditLogEntries();
  }

  function saveOAuth2Client(): void {
    logger.debug(`saveOAuth2Client called`);

    if (id === 0) {
      throw new Error('id cannot be zero!');
    } else if (!needsToBeSaved) {
      throw new Error('no changes to save!');
    }

    V1APIClient.saveOAuth2Client(oauth2Client)
      .then((response: AxiosResponse<OAuth2Client>) => {
        oauth2Client = { ...response.data };
        originalOAuth2Client = { ...response.data };
        needsToBeSaved = false;
      })
      .catch((error: AxiosError) => {
        if (error.response) {
          if (error.response.data) {
            oauth2ClientRetrievalError = error.response.data;
          }
        }
      });
  }

  function deleteOAuth2Client(): void {
    logger.debug(`deleteOAuth2Client called`);

    if (id === 0) {
      throw new Error('id cannot be zero!');
    }

    V1APIClient.deleteOAuth2Client(id)
      .then((response: AxiosResponse<OAuth2Client>) => {
        if (response.status === 204) {
          logger.debug(
            `navigating to /things/oauth2Clients because via deletion promise resolution`,
          );
          navigate('/things/oauth2Clients', { state: {}, replace: true });
        }
      })
      .catch((error: AxiosError) => {
        if (error.response) {
          if (error.response.data) {
            oauth2ClientRetrievalError = error.response.data;
          }
        }
      });
  }

  function fetchAuditLogEntries(): void {
    logger.debug(`deleteOAuth2Client called`);

    if (id === 0) {
      throw new Error('id cannot be zero!');
    }

    V1APIClient.fetchAuditLogEntriesForOAuth2Client(id)
      .then((response: AxiosResponse<AuditLogEntry[]>) => {
        auditLogEntries = response.data;
        logger.withValue('entries', auditLogEntries).debug('entries fetched');
      })
      .catch((error: AxiosError) => {
        if (error.response) {
          if (error.response.data) {
            oauth2ClientRetrievalError = error.response.data;
          }
        }
      });
  }
</script>

<div>
  <div
    class="relative flex flex-col min-w-0 break-words bg-white w-full mb-6 shadow-lg rounded">
    <div class="rounded-t mb-0 px-4 py-3 bg-transparent justify-between ">
      <div class="flex flex-wrap oauth2Clients-center">
        <div class="relative w-full max-w-full flex-grow flex-1">
          {#if originalOAuth2Client.id !== 0}
            <h2 class="text-gray-800 text-xl font-semibold">
              #{originalOAuth2Client.id}:
              {originalOAuth2Client.name}
            </h2>
          {/if}
        </div>
        <div class="flex w-full max-w-full flex-grow justify-end flex-1">
          <button
            class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded {needsToBeSaved ? '' : 'opacity-50 cursor-not-allowed'}"
            on:click={saveOAuth2Client}><i class="fa fa-save" />
            Save</button>
        </div>
      </div>
    </div>
    <div>
      <div class="flex flex-wrap -mx-3 mb-6">
        <div class="w-full md:w-1/2 px-3 mb-6 md:mb-0">
          <label
            class="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2"
            for="grid-first-name">
            {translationsToUse.labels.name}
          </label>
          <input
            class="appearance-none block w-full text-gray-700 border border-red-500 rounded py-3 px-4 mb-3 leading-tight focus:outline-none focus:bg-white"
            id="grid-first-name"
            type="text"
            on:keyup={evaluateChanges}
            bind:value={oauth2Client.name} />
          <!--  <p class="text-red-500 text-xs italic">Please fill out this field.</p>-->
        </div>
        <div class="w-full md:w-1/2 px-3">
          <label class="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2" for="grid-last-name">
            {translationsToUse.labels.details}
          </label>
          <input class="appearance-none block w-full text-gray-700 border border-gray-200 rounded py-3 px-4 leading-tight focus:outline-none focus:bg-white focus:border-gray-500" id="grid-last-name" type="text" on:keyup={evaluateChanges} bind:value={oauth2Client.details}>
        </div>
        <div
          class="flex w-full mr-3 mt-4 max-w-full flex-grow justify-end flex-1">
          <button
            class="bg-red-500 hover:bg-red-700 text-white font-bold py-2 px-4 rounded"
            on:click={deleteOAuth2Client}><i class="fa fa-trash-alt" />
            Delete</button>
        </div>
      </div>
    </div>
  </div>

  {#if currentUserStatus.isAdmin}
    <AuditLogTable entries={auditLogEntries} />
  {/if}
</div>
