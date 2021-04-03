<script lang="typescript">
// core components
import type { AxiosError, AxiosResponse } from 'axios';

import {
  ErrorResponse,
  fakeAPIClientFactory,
  APIClient,
  APIClientList,
  QueryFilter,
  UserSiteSettings,
  UserStatus,
} from '../../types';
import { Logger } from '../../logger';
import { frontendRoutes, statusCodes } from '../../constants';
import { V1APIClient } from '../../apiClient';

import APITable from '../APITable/APITable.svelte';
import { Superstore } from '../../stores';

export let location;

let queryFilter = QueryFilter.fromURLSearchParams();

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
  'src/views/APIClients.svelte',
);

let apiTableIncrementDisabled: boolean = false;
let apiTableDecrementDisabled: boolean = false;
let apiTableSearchQuery: string = '';

function searchAPIClients() {
  logger.debug('searchAPIClients called');

  if (superstore.frontendOnlyMode) {
    apiClients = fakeAPIClientFactory.buildList(10);
  } else {
    V1APIClient.searchForAPIClients(apiTableSearchQuery, queryFilter, adminMode)
      .then((response: AxiosResponse<APIClientList>) => {
        apiClients = response.data.apiClients || [];
        queryFilter.page -= 1;
      })
      .catch((error: AxiosError) => {
        apiClientRetrievalError = error.response?.data;
      });
  }
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
        apiClients = response.data.apiClients || [];

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
          if (error?.response?.data) {
            apiClientRetrievalError = error.response.data.message;
          }
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
      individualPageLink={frontendRoutes.LIST_ITEMS}
      creationLink='/things/apiClient'
      dataRetrievalError="{apiClientRetrievalError}"
      searchEnabled="{true}"
      searchFunction="{searchAPIClients}"
      incrementDisabled="{apiTableIncrementDisabled}"
      decrementDisabled="{apiTableDecrementDisabled}"
      incrementPageFunction="{incrementPage}"
      decrementPageFunction="{decrementPage}"
      fetchFunction="{fetchAPIClients}"
      deleteEnabled="{true}"
      deleteFunction="{promptDelete}"
      rowRenderFunction="{APIClient.asRow}"
    />
  </div>
</div>
