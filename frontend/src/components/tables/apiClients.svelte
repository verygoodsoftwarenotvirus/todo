<script lang="typescript">
// core components
import type { AxiosError, AxiosResponse } from 'axios';
import { onMount } from 'svelte';

import {
  ErrorResponse,
  APIClient,
  APIClientList,
  QueryFilter,
  UserSiteSettings,
  UserStatus,
  fakeUserFactory,
  fakeAPIClientFactory,
} from '../../types';
import { Logger } from '../../logger';
import { V1APIClient } from '../../apiClient';

import APITable from '../core/apiTable/apiTable.svelte';
import {frontendRoutes, statusCodes} from '../../constants';
import { Superstore } from '../../stores';

export let location;

let apiClientRetrievalError = '';
let apiClients: APIClient[] = [];

let adminMode: boolean = false;
let currentAuthStatus: UserStatus = new UserStatus();
let currentSessionSettings = new UserSiteSettings();
let translationsToUse = currentSessionSettings.getTranslations().models.apiClient;

let superstore = new Superstore({
  userStatusStoreUpdateFunc: (value: UserStatus) => {
    currentAuthStatus = value;
  },
  sessionSettingsStoreUpdateFunc: (value: UserSiteSettings) => {
    currentSessionSettings = value;
    translationsToUse = currentSessionSettings.getTranslations().models.apiClient;
  },
  adminModeUpdateFunc: (value: boolean) => {
    adminMode = value;
  },
});

let logger = new Logger().withDebugValue(
  'source',
  'src/views/admin/apiClients.svelte',
);

onMount(fetchAPIClients);

// begin experimental API table code

let queryFilter = QueryFilter.fromURLSearchParams();

let apiTableIncrementDisabled: boolean = false;
let apiTableDecrementDisabled: boolean = false;
let apiTableSearchQuery: string = '';

function searchAPIClients() {
  logger.debug('searchAPIClients called');
}

function incrementPage() {
  if (!apiTableIncrementDisabled) {
    logger.debug(`incrementPage called`);
    queryFilter.page += 1;
    fetchAPIClients();
  }
}

function decrementPage() {
  if (!apiTableDecrementDisabled) {
    logger.debug(`decrementPage called`);
    queryFilter.page -= 1;
    fetchAPIClients();
  }
}

function fetchAPIClients() {
  logger.debug('fetchAPIClients called');

  if (superstore.frontendOnlyMode) {
    apiClients = fakeAPIClientFactory.buildList(queryFilter.limit);
  } else {
    V1APIClient.fetchListOfAPIClients(queryFilter, adminMode)
      .then((response: AxiosResponse<APIClientList>) => {
        apiClients = response.data.clients || [];

        console.dir(apiClients)

        queryFilter.page = response.data.page;
        apiTableIncrementDisabled = apiClients.length === 0;
        apiTableDecrementDisabled = queryFilter.page === 1;
      })
      .catch((error: AxiosError) => {
        apiClientRetrievalError = error.response?.data;
      });
  }
}

function promptDelete(id: number) {
  logger.debug('promptDelete called');

  if (confirm(`are you sure you want to delete apiClient #${id}?`)) {
    if (superstore.frontendOnlyMode) {
      fetchAPIClients();
    } else {
      V1APIClient.deleteAPIClient(id)
        .then((response: AxiosResponse<APIClient>) => {
          if (response?.status === statusCodes.NO_CONTENT) {
            fetchAPIClients();
          }
        })
        .catch((error: AxiosError<ErrorResponse>) => {
          apiClientRetrievalError = error.response?.data?.message || '';
        });
    }
  }
}
</script>

<div class="flex flex-wrap mt-4">
  <div class="w-full mb-12 px-4">
    <APITable
      title="APIClients"
      headers="{APIClient.headers(translationsToUse)}"
      rows="{apiClients}"
      creationLink="{frontendRoutes.CREATE_API_CLIENT}"
      individualPageLink="{frontendRoutes.INDIVIDUAL_API_CLIENT}"
      dataRetrievalError="{apiClientRetrievalError}"
      searchFunction="{searchAPIClients}"
      incrementDisabled="{apiTableIncrementDisabled}"
      decrementDisabled="{apiTableDecrementDisabled}"
      incrementPageFunction="{incrementPage}"
      decrementPageFunction="{decrementPage}"
      fetchFunction="{fetchAPIClients}"
      deleteFunction="{promptDelete}"
      rowRenderFunction="{APIClient.asRow}"
    />
  </div>
</div>
