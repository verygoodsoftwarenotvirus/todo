<script lang="typescript">
  import axios, { AxiosResponse, AxiosError } from "axios";
  import { onMount, onDestroy } from "svelte";
  import { link, navigate } from "svelte-routing";

  import { renderUnixTime, inheritQueryFilterSearchParams } from "../../../utils"
  import { User, UserList } from "../../../models"

  let searchQuery: string = '';
  let currentPage: number = 0;
  let pageQuantity: number = 20;

  // these should be overridden by the initial fetch
  let incrementDisabled: boolean = true;
  let decrementDisabled: boolean = true;

  export let users: User[] = [];
  export let dataRetrievalError: string = '';

  import {Logger} from "../../../logger";
  let logger = new Logger().withDebugValue("source", "src/components/Things/Tables/AdminUsersTable.svelte");

  import { adminModeStore } from "../../../stores";
  let adminMode: boolean = false;
  const unsubscribeFromAdminModeUpdates = adminModeStore.subscribe((value: boolean) => {
    adminMode = value;
    fetchUsers();
  });

  import { authStatusStore } from "../../../stores";
  let currentAuthStatus = {};
  const unsubscribeFromAuthStatusUpdates = authStatusStore.subscribe((value: UserStatus) => {
    currentAuthStatus = value;
  });
  // onDestroy(unsubscribeFromAuthStatusUpdates);

  function search(): void {
    if (searchQuery.length >= 3) {
      logger.debug(`searching for items: ${searchQuery}`)
      searchUsers();
    }
  }

  function searchUsers() {
    logger.debug("searchUsers called");

    throw new Error("unimplemented");

    const path: string = "/api/v1/users/search";

    // const pageURLParams: URLSearchParams = new URLSearchParams(window.location.search);
    // const outboundURLParams: URLSearchParams = inheritQueryFilterSearchParams(pageURLParams);
    //
    // if (adminMode) {
    //   outboundURLParams.set("admin", "true");
    // }
    // outboundURLParams.set("q", searchQuery)
    //
    // const qs = outboundURLParams.toString()
    // const uri = `${path}?${qs}`;
    //
    // axios.get(uri, { withCredentials: true })
    //         .then((response: AxiosResponse<UserList>) => {
    //           items = response.data || [];
    //           currentPage = -1;
    //         })
    //         .catch((error: AxiosError) => {
    //           if (error.response) {
    //             if (error.response.data) {
    //               itemRetrievalError = error.response.data;
    //             }
    //           }
    //         });
  }

  function incrementPage() {
    if (!incrementDisabled) {
        logger.debug(`incrementPage called`);
        currentPage += 1;
        fetchUsers();
    }
  }

  function decrementPage() {
    if (!decrementDisabled) {
        logger.debug(`decrementPage called`);
        currentPage -= 1;
        fetchUsers();
    }
  }

  function examplePostDeletionFunc() {
    logger.debug(`examplePostDeletionFunc called`);
  }

  function fetchUsers() {
    logger.debug("fetchUsers called");

    const path: string = "/api/v1/users";

    const pageURLParams: URLSearchParams = new URLSearchParams(window.location.search);
    const outboundURLParams: URLSearchParams = inheritQueryFilterSearchParams(pageURLParams);

    if (adminMode) {
      outboundURLParams.set("admin", "true");
    }
    outboundURLParams.set("page", currentPage.toString());
    outboundURLParams.set("limit", pageQuantity.toString());

    const qs = outboundURLParams.toString()
    const uri = `${path}?${qs}`;

    axios.get(uri, { withCredentials: true })
            .then((response: AxiosResponse<UserList>) => {
              users = response.data.users || [];

              console.dir(users);

              currentPage = response.data.page;
              incrementDisabled = users.length === 0;
              decrementDisabled = currentPage === 1;
            })
            .catch((error: AxiosError) => {
              if (error.response) {
                if (error.response.data) {
                  dataRetrievalError = error.response.data;
                }
              }
            });
  }

  function promptDelete(id: number) {
    logger.debug("promptDelete called");

    if (confirm(`are you sure you want to delete user #${id}?`)) {
      const path: string = `/api/v1/users/${id}`;

      axios.delete(path, { withCredentials: true })
              .then((response: AxiosResponse<User>) => {
                if (response.status === 204) {
                  fetchUsers();
                }
              })
              .catch((error: AxiosError) => {
                if (error.response) {
                  if (error.response.data) {
                    dataRetrievalError = error.response.data;
                  }
                }
              });
    }
  }
</script>

<div class="relative flex flex-col min-w-0 break-words w-full mb-6 shadow-lg rounded bg-white">
  <div class="rounded-t mb-0 px-4 py-3 border-0">
    <div class="flex flex-wrap items-center">
      <div class="relative w-full px-4 max-w-full flex-grow flex-1">
        <h3 class="font-semibold text-lg text-gray-800">
          Users
          <button class="border-2 font-bold py-1 px-4 m-2 rounded" on:click={fetchUsers}>
            🔄
          </button>
        </h3>
      </div>

      <div class="text-center">
        <div class="px-4 py-2 m-2">
          <button on:click={decrementPage} disabled={decrementDisabled}><i class="fa fa-arrow-circle-left"></i></button>
          &nbsp;
          {#if currentPage > 0 }
            Page {currentPage}
          {/if}
          &nbsp;
          <button on:click={incrementPage} disabled={incrementDisabled}><i class="fa fa-arrow-circle-right"></i></button>
        </div>
      </div>

      <div>
        per page:
        <select bind:value={pageQuantity} on:blur={fetchUsers} class="appearance-none border p-1 rounded leading-tight">
          <option>20</option>
          <option>35</option>
          <option>50</option>
          <option>100</option>
          <option>200</option>
        </select>
      </div>

      <span class="mr-2 ml-2">

      {#if dataRetrievalError !== ''}
        <span class="text-red-600">{dataRetrievalError}</span>
      {/if}

      <span class="mr-2 ml-2">

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
          <th class="px-6 align-middle border border-solid py-3 text-xs uppercase border-l-0 border-r-0 whitespace-no-wrap font-semibold text-left bg-gray-100 text-gray-600 border-gray-200">
            ID
          </th>
          <th class="px-6 align-middle border border-solid py-3 text-xs uppercase border-l-0 border-r-0 whitespace-no-wrap font-semibold text-left bg-gray-100 text-gray-600 border-gray-200">
            Username
          </th>
          <th class="px-6 align-middle border border-solid py-3 text-xs uppercase border-l-0 border-r-0 whitespace-no-wrap font-semibold text-left bg-gray-100 text-gray-600 border-gray-200">
            Due For Password Change?
          </th>
          <th class="px-6 align-middle border border-solid py-3 text-xs uppercase border-l-0 border-r-0 whitespace-no-wrap font-semibold text-left bg-gray-100 text-gray-600 border-gray-200">
            Password Last Changed
          </th>
          <th class="px-6 align-middle border border-solid py-3 text-xs uppercase border-l-0 border-r-0 whitespace-no-wrap font-semibold text-left bg-gray-100 text-gray-600 border-gray-200">
            Admin?
          </th>
          <th class="px-6 align-middle border border-solid py-3 text-xs uppercase border-l-0 border-r-0 whitespace-no-wrap font-semibold text-left bg-gray-100 text-gray-600 border-gray-200">
            Created On
          </th>
          <th class="px-6 align-middle border border-solid py-3 text-xs uppercase border-l-0 border-r-0 whitespace-no-wrap font-semibold text-left bg-gray-100 text-gray-600 border-gray-200">
            Last Updated On
          </th>
          <th class="px-6 align-middle border border-solid py-3 text-xs uppercase border-l-0 border-r-0 whitespace-no-wrap font-semibold text-left bg-gray-100 text-gray-600 border-gray-200">
            Delete
          </th>
        </tr>
      </thead>
      <tbody>
      {#each users as user}
        <tr>
          <a use:link href="/admin/users/{user.id}">
            <th class="border-t-0 px-6 align-middle border-l-0 border-r-0 text-xs whitespace-no-wrap p-4 text-left flex items-center">
              <span class="ml-3 font-bold btext-gray-700">
                {user.id}
              </span>
            </th>
          </a>
          <td class="border-t-0 px-6 align-middle border-l-0 border-r-0 text-xs whitespace-no-wrap p-4">
            {user.username}
          </td>
          <td class="border-t-0 px-6 align-middle border-l-0 border-r-0 text-xs whitespace-no-wrap p-4">
            {user.requiresPasswordChange}
          </td>
          <td class="border-t-0 px-6 align-middle border-l-0 border-r-0 text-xs whitespace-no-wrap p-4">
            {renderUnixTime(user.passwordLastChangedOn)}
          </td>
          <td class="border-t-0 px-6 align-middle border-l-0 border-r-0 text-xs whitespace-no-wrap p-4">
            {user.isAdmin}
          </td>
          <td class="border-t-0 px-6 align-middle border-l-0 border-r-0 text-xs whitespace-no-wrap p-4">
            {renderUnixTime(user.createdOn)}
          </td>
          <td class="border-t-0 px-6 align-middle border-l-0 border-r-0 text-xs whitespace-no-wrap p-4">
            {renderUnixTime(user.lastUpdatedOn)}
          </td>
          <td class="border-t-0 px-6 align-middle border-l-0 border-r-0 text-xs whitespace-no-wrap p-4 text-right text-red-600">
            <div on:click={promptDelete(user.id)}><i class="fa fa-trash"></i></div>
          </td>
        </tr>
      {/each}
      </tbody>
    </table>
  </div>
</div>
