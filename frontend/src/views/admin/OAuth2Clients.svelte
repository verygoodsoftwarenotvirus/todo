<script lang="typescript">
// core components
import { AxiosError, AxiosResponse } from 'axios';
import { onMount } from 'svelte';

import {
  ErrorResponse,
  fakeOAuth2ClientFactory,
  fakeUserFactory,
  OAuth2Client,
  OAuth2ClientList,
  QueryFilter,
  UserSiteSettings,
  UserStatus,
} from '../../types';
import { Logger } from '../../logger';
import { V1APIClient } from '../../apiClient';

import APITable from '../../components/APITable/APITable.svelte';
import { statusCodes } from '../../constants';
import { Superstore } from '../../stores';

export let location;

let oauth2ClientRetrievalError = '';
let oauth2Clients: OAuth2Client[] = [];

let adminMode: boolean = false;
let currentAuthStatus: UserStatus = new UserStatus();
let currentSessionSettings = new UserSiteSettings();
let translationsToUse = currentSessionSettings.getTranslations().models
  .oauth2Client;

let superstore = new Superstore({
  userStatusStoreUpdateFunc: (value: UserStatus) => {
    currentAuthStatus = value;
  },
  sessionSettingsStoreUpdateFunc: (value: UserSiteSettings) => {
    currentSessionSettings = value;
    translationsToUse = currentSessionSettings.getTranslations().models
      .oauth2Client;
  },
  adminModeUpdateFunc: (value: boolean) => {
    adminMode = value;
  },
});

let logger = new Logger().withDebugValue(
  'source',
  'src/views/admin/OAuth2Clients.svelte',
);

onMount(fetchOAuth2Clients);

// begin experimental API table code

let queryFilter = QueryFilter.fromURLSearchParams();

let apiTableIncrementDisabled: boolean = false;
let apiTableDecrementDisabled: boolean = false;
let apiTableSearchQuery: string = '';

function searchOAuth2Clients() {
  logger.debug('searchOAuth2Clients called');
}

function incrementPage() {
  if (!apiTableIncrementDisabled) {
    logger.debug(`incrementPage called`);
    queryFilter.page += 1;
    fetchOAuth2Clients();
  }
}

function decrementPage() {
  if (!apiTableDecrementDisabled) {
    logger.debug(`decrementPage called`);
    queryFilter.page -= 1;
    fetchOAuth2Clients();
  }
}

function fetchOAuth2Clients() {
  logger.debug('fetchOAuth2Clients called');

  if (superstore.frontendOnlyMode) {
    oauth2Clients = fakeOAuth2ClientFactory.buildList(queryFilter.limit);
  } else {
    V1APIClient.fetchListOfOAuth2Clients(queryFilter, adminMode)
      .then((response: AxiosResponse<OAuth2ClientList>) => {
        oauth2Clients = response.data.clients || [];

        queryFilter.page = response.data.page;
        apiTableIncrementDisabled = oauth2Clients.length === 0;
        apiTableDecrementDisabled = queryFilter.page === 1;
      })
      .catch((error: AxiosError) => {
        oauth2ClientRetrievalError = error.response?.data;
      });
  }
}

function promptDelete(id: number) {
  logger.debug('promptDelete called');

  if (confirm(`are you sure you want to delete oauth2Client #${id}?`)) {
    const path: string = `/api/v1/oauth2Clients/${id}`;

    V1APIClient.deleteOAuth2Client(id)
      .then((response: AxiosResponse<OAuth2Client>) => {
        if (response.status === statusCodes.NO_CONTENT) {
          fetchOAuth2Clients();
        }
      })
      .catch((error: AxiosError<ErrorResponse>) => {
        if (error.response) {
          if (error.response.data) {
            oauth2ClientRetrievalError = error.response.data.message;
          }
        }
      });
  }
}
</script>

<div class="flex flex-wrap mt-4">
  <div class="w-full mb-12 px-4">
    <APITable
      title="OAuth2Clients"
      headers="{OAuth2Client.headers(translationsToUse)}"
      rows="{oauth2Clients}"
      individualPageLink="/admin/oauth2_clients"
      newPageLink="/admin/oauth2_clients/new"
      dataRetrievalError="{oauth2ClientRetrievalError}"
      searchFunction="{searchOAuth2Clients}"
      incrementDisabled="{apiTableIncrementDisabled}"
      decrementDisabled="{apiTableDecrementDisabled}"
      incrementPageFunction="{incrementPage}"
      decrementPageFunction="{decrementPage}"
      fetchFunction="{fetchOAuth2Clients}"
      deleteFunction="{promptDelete}"
      rowRenderFunction="{OAuth2Client.asRow}"
    />
  </div>
</div>
