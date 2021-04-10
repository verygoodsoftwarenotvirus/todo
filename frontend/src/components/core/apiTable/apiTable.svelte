<script lang="typescript">
import { link, navigate } from 'svelte-routing';
import JSONTree from 'svelte-json-tree';

import { Logger } from '../../../logger';
import { QueryFilter, UserSiteSettings, UserStatus } from '../../../types';
import { Superstore } from '../../../stores';
import type { APITableCell, APITableHeader } from './types';

let logger = new Logger().withDebugValue(
  'source',
  'src/components/apiTable/apiTable.svelte',
);

interface DatabaseRecord {
  id: number;
}

// local state
let searchQuery: string = '';
let currentPage: number = 0;
export let dataRetrievalError: string = '';

export let title: string = '';
export let headers: APITableHeader[] = [];
export let rows: DatabaseRecord[] = [];

export let queryFilter: QueryFilter = new QueryFilter();

export let creationLink: string = '';
export let individualPageLink: string = '';

export let searchEnabled: boolean = true;
export let searchFunction: (((query: string) => void) | null);

export let deleteEnabled: boolean = true;
export let deleteFunction: (id: number) => void;

export let incrementDisabled: boolean = true;
export let incrementPageFunction: () => void;

export let decrementDisabled: boolean = true;
export let decrementPageFunction: () => void;

export let fetchFunction: () => void;
export let rowRenderFunction: (rowContent: any) => APITableCell[];

let adminMode: boolean = false;
let currentAuthStatus: UserStatus = new UserStatus();
let currentSessionSettings = new UserSiteSettings();
let translationsToUse = currentSessionSettings.getTranslations().components
  .apiTable;

let superstore = new Superstore({
  userStatusStoreUpdateFunc: (value: UserStatus) => {
    logger.withDebugValue("value", value).debug(`new UserStatus received in APITable`)
    currentAuthStatus = value;
  },
  sessionSettingsStoreUpdateFunc: (value: UserSiteSettings) => {
    currentSessionSettings = value;
    translationsToUse = currentSessionSettings.getTranslations().components.apiTable;
  },
  adminModeUpdateFunc: (value: boolean) => {
    adminMode = value;
    fetchFunction();
  },
});

function search(): void {
  if (searchQuery.length >= 3) {
    logger.debug(`searching for: ${searchQuery}`);
    if (searchEnabled && searchFunction !== null) {
      searchFunction(searchQuery);
    }
  }
}

function goToNewPage() {
  logger.debug(`navigating to ${creationLink} via goToNewPage`);
  navigate(creationLink, { state: {}, replace: true });
}
</script>

<div class="relative flex flex-col min-w-0 break-words w-full mb-6 shadow-lg rounded bg-white">
  <div class="rounded-t mb-0 px-4 py-3 border-0">
    <div class="flex flex-wrap items-center">
      <div class="relative w-full px-4 max-w-full flex-grow flex-1">
        <h3 class="font-semibold text-lg text-gray-800">
          {title}

          {#if goToNewPage !== undefined && creationLink !== ''}
            <button
              class="border-2 font-bold py-1 px-4 m-2 rounded"
              on:click="{goToNewPage}"
            >
              🆕
            </button>
          {/if}

          {#if fetchFunction !== undefined}
            <button
              class="border-2 font-bold py-1 px-4 m-2 rounded"
              on:click="{fetchFunction}"
            >
              🔄
            </button>
          {/if}
        </h3>
      </div>

      <div class="text-center">
        <div class="px-4 py-2 m-2">
          {#if decrementPageFunction !== undefined}
            <button
              on:click="{decrementPageFunction}"
              disabled="{decrementDisabled}"
            ><i class="fa fa-arrow-circle-left"></i></button>
          {/if}
          &nbsp;
          {#if currentPage > 0}{translationsToUse.page} {currentPage}{/if}
          &nbsp;
          {#if incrementPageFunction !== undefined}
            <button
              on:click="{incrementPageFunction}"
              disabled="{incrementDisabled}"
            ><i class="fa fa-arrow-circle-right"></i></button>
          {/if}
        </div>
      </div>

      <div>
        {translationsToUse.perPage}:
        <select
          bind:value="{queryFilter.limit}"
          on:blur="{fetchFunction}"
          class="appearance-none border p-1 rounded leading-tight"
        >
          <option>20</option>
          <option>35</option>
          <option>50</option>
          <option>100</option>
          <option>200</option>
        </select>
      </div>

      <span class="mr-2 ml-2"></span>

      {#if dataRetrievalError !== ''}
        <span class="text-red-600">{dataRetrievalError}</span>
      {/if}

      <span class="mr-2 ml-2"></span>

      {#if searchEnabled}
        <div class="flex border-grey-light border">
          <input
            class="w-full rounded ml-1"
            type="text"
            placeholder="{translationsToUse.inputPlaceholders.search}"
            bind:value="{searchQuery}"
            on:keyup="{search}"
          />
        </div>
      {/if}
    </div>
  </div>
  <div class="block w-full overflow-x-auto">
    <table class="items-center w-full bg-transparent border-collapse">
      <thead>
        <tr>
          {#each headers as header}
            {#if header.requiresAdmin}
              {#if currentAuthStatus.adminPermissions !== null && adminMode}
                <th
                  class="px-6 align-middle border border-solid py-3 text-xs uppercase border-l-0 border-r-0 whitespace-no-wrap font-semibold text-left bg-gray-100 text-gray-600 border-gray-200"
                >
                  {header.content}
                </th>
              {/if}
            {:else}
              <th
                class="px-6 align-middle border border-solid py-3 text-xs uppercase border-l-0 border-r-0 whitespace-no-wrap font-semibold text-left bg-gray-100 text-gray-600 border-gray-200"
              >
                {header.content}
              </th>
            {/if}
          {/each}
          {#if deleteEnabled}
            <th
              class="px-6 align-middle border border-solid py-3 text-xs uppercase border-l-0 border-r-0 whitespace-no-wrap font-semibold text-left bg-gray-100 text-gray-600 border-gray-200"
            >
              {translationsToUse.delete}
            </th>
          {/if}
        </tr>
      </thead>
      <tbody>
        {#each rows as row}
          <tr>
            {#each rowRenderFunction(row) as cell}
              {#if cell.isIDCell && individualPageLink !== ''}
                <a use:link href="{individualPageLink}/{row.id}">
                  <th class="border-t-0 px-6 align-middle border-l-0 border-r-0 text-xs whitespace-no-wrap p-4 text-left flex items-center">
                    <span class="ml-3 font-bold btext-gray-700">
                      {row.id}
                    </span>
                  </th>
                </a>
              {:else if cell.requiresAdmin}
                {#if currentAuthStatus.adminPermissions !== null && adminMode}
                  {#if cell.isJSON}
                    <JSONTree value="{JSON.parse(cell.content)}" />
                  {:else}
                    <td class="border-t-0 px-6 align-middle border-l-0 border-r-0 text-xs whitespace-no-wrap p-4" >
                      {cell.content}
                    </td>
                  {/if}
                {/if}
              {:else if cell.isJSON}
                <JSONTree value="{JSON.parse(cell.content)}" />
              {:else}
                <td class="border-t-0 px-6 align-middle border-l-0 border-r-0 text-xs whitespace-no-wrap p-4" >
                  {cell.content}
                </td>
              {/if}
            {/each}
            {#if deleteFunction !== undefined && deleteEnabled}
              <td
                class="border-t-0 px-6 align-middle border-l-0 border-r-0 text-xs whitespace-no-wrap p-4 text-right text-red-600"
                on:click="{deleteFunction(row.id)}"
              >
                <div><i class="fa fa-trash"></i></div>
              </td>
            {/if}
          </tr>
        {/each}

        <!--{#if currentAuthStatus.adminPermissions !== null && adminMode}-->
        <!--  <td class="border-t-0 px-6 align-middle border-l-0 border-r-0 text-xs whitespace-no-wrap p-4">-->
        <!--    {item.belongsToUser}-->
        <!--  </td>-->
        <!--{/if}-->
      </tbody>
    </table>
  </div>
</div>
