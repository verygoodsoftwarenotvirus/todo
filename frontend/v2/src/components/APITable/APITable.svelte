<script lang="typescript">
  import { link, navigate } from "svelte-routing";

  import { UserStatus } from "../../models"

  // local state
  let searchQuery: string = '';
  let currentPage: number = 0;
  let pageQuantity: number = 20;

  export let title: string = '';
  export let headers: string[] = [];
  export let rows: string[][] = [[]];
  export let newPageLink: string = '';
  export let individualPageLink: string= '';

  export let dataRetrievalError: string = '';

  export let searchFunction;
  export let incrementDisabled;
  export let incrementPageFunction;
  export let decrementDisabled;
  export let decrementPageFunction;
  export let fetchFunction;
  export let deleteFunction;
  export let rowRenderFunction;

  import {Logger} from "../../logger";
  let logger = new Logger().withDebugValue("source", "src/components/Things/Tables/APITable.svelte");

  import { adminModeStore } from "../../stores";
  let adminMode: boolean = false;
  const unsubscribeFromAdminModeUpdates = adminModeStore.subscribe((value: boolean) => {
    adminMode = value;
    fetchFunction();
  });

  import { authStatusStore } from "../../stores";
  let currentAuthStatus = {};
  const unsubscribeFromAuthStatusUpdates = authStatusStore.subscribe((value: UserStatus) => {
    currentAuthStatus = value;
  });
  // onDestroy(unsubscribeFromAuthStatusUpdates);

  function search(): void {
    if (searchQuery.length >= 3) {
      logger.debug(`searching for items: ${searchQuery}`);
      searchFunction();
    }
  }

  function goToNewPage() {
    logger.debug(`navigating to ${newPageLink} via goToNewPage`);
    navigate(newPageLink, { state: {}, replace: true });
  }
</script>

<div class="relative flex flex-col min-w-0 break-words w-full mb-6 shadow-lg rounded bg-white">
  <div class="rounded-t mb-0 px-4 py-3 border-0">
    <div class="flex flex-wrap items-center">
      <div class="relative w-full px-4 max-w-full flex-grow flex-1">
        <h3 class="font-semibold text-lg text-gray-800">
          {title}
          <button class="border-2 font-bold py-1 px-4 m-2 rounded" on:click={goToNewPage}>
            🆕
          </button>
          <button class="border-2 font-bold py-1 px-4 m-2 rounded" on:click={fetchFunction}>
            🔄
          </button>
        </h3>
      </div>

      <div class="text-center">
        <div class="px-4 py-2 m-2">
          <button on:click={decrementPageFunction} disabled={decrementDisabled}><i class="fa fa-arrow-circle-left"></i></button>
          &nbsp;
          {#if currentPage > 0 }
            Page {currentPage}
          {/if}
          &nbsp;
          <button on:click={incrementPageFunction} disabled={incrementDisabled}><i class="fa fa-arrow-circle-right"></i></button>
        </div>
      </div>

      <div>
        per page:
        <select bind:value={pageQuantity} on:blur={fetchFunction} class="appearance-none border p-1 rounded leading-tight">
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

      <div class="flex border-grey-light border">
        <input class="w-full rounded ml-1" type="text" placeholder="Search..." bind:value={searchQuery} on:keyup={search}>
      </div>
    </div>
  </div>
  <div class="block w-full overflow-x-auto">
    <!-- Projects table -->
    <table class="items-center w-full bg-transparent border-collapse">
      <thead>
        <tr>
          {#each headers as header}
            {#if header.requiresAdmin}
              {#if currentAuthStatus.isAdmin && adminMode}
                <th class="px-6 align-middle border border-solid py-3 text-xs uppercase border-l-0 border-r-0 whitespace-no-wrap font-semibold text-left bg-gray-100 text-gray-600 border-gray-200">
                  {header.content}
                </th>
              {/if}
            {:else}
              <th class="px-6 align-middle border border-solid py-3 text-xs uppercase border-l-0 border-r-0 whitespace-no-wrap font-semibold text-left bg-gray-100 text-gray-600 border-gray-200">
                {header.content}
              </th>
            {/if}
          {/each}
          <th class="px-6 align-middle border border-solid py-3 text-xs uppercase border-l-0 border-r-0 whitespace-no-wrap font-semibold text-left bg-gray-100 text-gray-600 border-gray-200">
            Delete
          </th>
        </tr>
      </thead>
      <tbody>

      {#each rows as row}
        <tr>
          {#each rowRenderFunction(row) as cell}
            {#if cell.fieldName === 'id'}
              <a use:link href="{individualPageLink}/{row.id}">
                <th class="border-t-0 px-6 align-middle border-l-0 border-r-0 text-xs whitespace-no-wrap p-4 text-left flex items-center">
                  <span class="ml-3 font-bold btext-gray-700">
                    {row.id}
                  </span>
                </th>
              </a>
            {:else}
              {#if cell.requiresAdmin}
                {#if currentAuthStatus.isAdmin && adminMode}
                  <td class="border-t-0 px-6 align-middle border-l-0 border-r-0 text-xs whitespace-no-wrap p-4">
                    {cell.content}
                  </td>
                {/if}
              {:else}
                <td class="border-t-0 px-6 align-middle border-l-0 border-r-0 text-xs whitespace-no-wrap p-4">
                  {cell.content}
                </td>
              {/if}
            {/if}
          {/each}
          <td class="border-t-0 px-6 align-middle border-l-0 border-r-0 text-xs whitespace-no-wrap p-4 text-right text-red-600" on:click={deleteFunction(row.id)}>
            <div><i class="fa fa-trash"></i></div>
          </td>
        </tr>
      {/each}

        <!--{#if currentAuthStatus.isAdmin && adminMode}-->
        <!--  <td class="border-t-0 px-6 align-middle border-l-0 border-r-0 text-xs whitespace-no-wrap p-4">-->
        <!--    {item.belongsToUser}-->
        <!--  </td>-->
        <!--{/if}-->

      </tbody>
    </table>
  </div>
</div>
