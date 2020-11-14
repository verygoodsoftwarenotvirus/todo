<script lang="typescript">
  import { AxiosError, AxiosResponse } from 'axios';
  import JSONTree from 'svelte-json-tree';

  import { AuditLogEntry, QueryFilter, UserSiteSettings } from '../../types';
  import { renderUnixTime } from '../../utils';
  import { Logger } from '../../logger';
  import { translations } from '../../i18n';
  import { sessionSettingsStore } from '../../stores';
  import { onDestroy } from 'svelte';

  export let entries: AuditLogEntry[] = [];
  export let entryFetchFunc: Promise<AxiosResponse<AuditLogEntry[]>>;

  let searchQuery: string = '';
  let retrievalError: string = '';
  let queryFilter = new QueryFilter();
  let decrementDisabled = false;
  let incrementDisabled = false;

  let logger = new Logger().withDebugValue(
    'source',
    'src/components/AuditLogTable/AuditLogTable.svelte',
  );

  // set up translations
  let currentSessionSettings = new UserSiteSettings();
  let translationsToUse = translations.messagesFor(
    currentSessionSettings.language,
  ).components.auditLogEntryTable;
  const unsubscribeFromSettingsUpdates = sessionSettingsStore.subscribe(
    (value: UserSiteSettings) => {
      currentSessionSettings = value;
      translationsToUse = translations.messagesFor(
        currentSessionSettings.language,
      ).components.auditLogEntryTable;
    },
  );
  onDestroy(unsubscribeFromSettingsUpdates);

  function fetchEntries() {
    entryFetchFunc
      .then((response: AxiosResponse<AuditLogEntry[]>) => {
        entries = response.data;
        logger.withValue('entries', entries).debug('entries fetched');
      })
      .catch((error: AxiosError) => {
        if (error.response) {
          if (error.response.data) {
            retrievalError = error.response.data;
          }
        }
      });
  }
</script>

<div
  class="relative flex flex-col min-w-0 break-words bg-white w-full mb-6 shadow-lg rounded">
  <div class="rounded-t mb-0 px-4 py-3 bg-transparent justify-between ">
    <div class="flex flex-wrap items-center">
      <div class="relative w-full max-w-full flex-grow flex-1">
        <h2 class="text-gray-800 text-xl font-semibold">Audit Log</h2>
      </div>

      <div class="text-center">
        <div class="px-4 py-2 m-2">
          <button
            on:click={console.log('decrementPageFunction')}
            disabled={decrementDisabled}><i
              class="fa fa-arrow-circle-left" /></button>
          &nbsp;
          {#if queryFilter.page > 0}
            {'translationsToUse.page'}
            {queryFilter.page}
          {/if}
          &nbsp;
          <button
            on:click={console.log('incrementPageFunction')}
            disabled={incrementDisabled}><i
              class="fa fa-arrow-circle-right" /></button>
        </div>
      </div>

      <span class="mr-2 ml-2" />

      {#if retrievalError !== ''}
        <span class="text-red-600">{retrievalError}</span>
      {/if}

      <span class="mr-2 ml-2" />

      <div class="flex border-grey-light border">
        <input
          class="w-full rounded ml-1"
          type="text"
          placeholder={'translationsToUse.inputPlaceholders.search'}
          bind:value={searchQuery}
          on:keyup={console.log('search')} />
      </div>
    </div>
  </div>
  <div>
    <div class="flex flex-wrap -mx-3 mb-6">
      <table class="items-center w-full bg-transparent border-collapse">
        <thead>
          <tr>
            <th
              class="px-6 align-middle border border-solid py-3 text-xs uppercase border-l-0 border-r-0 whitespace-no-wrap font-semibold text-left bg-gray-100 text-gray-600 border-gray-200">
              ID
            </th>
            <th
              class="px-6 align-middle border border-solid py-3 text-xs uppercase border-l-0 border-r-0 whitespace-no-wrap font-semibold text-left bg-gray-100 text-gray-600 border-gray-200">
              Event Type
            </th>
            <th
              class="px-6 align-middle border border-solid py-3 text-xs uppercase border-l-0 border-r-0 whitespace-no-wrap font-semibold text-left bg-gray-100 text-gray-600 border-gray-200">
              Context
            </th>
            <th
              class="px-6 align-middle border border-solid py-3 text-xs uppercase border-l-0 border-r-0 whitespace-no-wrap font-semibold text-left bg-gray-100 text-gray-600 border-gray-200">
              Created On
            </th>
          </tr>

          <!--
            <tr>
              {#each AuditLogEntry.headers(translationsToUse) as header}
                <th class="px-6 align-middle border border-solid py-3 text-xs uppercase border-l-0 border-r-0 whitespace-no-wrap font-semibold text-left bg-gray-100 text-gray-600 border-gray-200">
                  {header.content}
                </th>
              {/each}
            </tr>
          -->
        </thead>
        <tbody>
        {#each entries as entry}
          <tr>
            <td class="border-t-0 px-6 align-middle border-l-0 border-r-0 text-xs whitespace-no-wrap p-4">
              {entry.id}
            </td>
            <td class="border-t-0 px-6 align-middle border-l-0 border-r-0 text-xs whitespace-no-wrap p-4">
              {entry.eventType}
            </td>
            <td class="border-t-0 px-6 align-middle border-l-0 border-r-0 text-xs whitespace-no-wrap p-4">
              <JSONTree value={entry.context}/>
            </td>
            <td class="border-t-0 px-6 align-middle border-l-0 border-r-0 text-xs whitespace-no-wrap p-4">
              {renderUnixTime(entry.createdOn)}
            </td>
          </tr>
        {/each}
        </tbody>
      </table>
    </div>
  </div>
</div>
