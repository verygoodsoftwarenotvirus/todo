<script lang="typescript">
// core components
import { AxiosError, AxiosResponse } from 'axios';
import { onMount } from 'svelte';

import {
  ErrorResponse,
  Webhook,
  WebhookList,
  QueryFilter,
  UserSiteSettings,
  UserStatus,
  fakeWebhookFactory,
} from '../../types';
import { Logger } from '../../logger';
import { V1APIClient } from '../../apiClient';

import APITable from '../core/apiTable/apiTable.svelte';
import { frontendRoutes, statusCodes } from '../../constants';
import { Superstore } from '../../stores';

export let location;

let webhookRetrievalError = '';
let webhooks: Webhook[] = [];

let adminMode: boolean = false;
let currentAuthStatus: UserStatus = new UserStatus();
let currentSessionSettings = new UserSiteSettings();
let translationsToUse = currentSessionSettings.getTranslations().models.webhook;

let superstore = new Superstore({
  userStatusStoreUpdateFunc: (value: UserStatus) => {
    currentAuthStatus = value;
  },
  sessionSettingsStoreUpdateFunc: (value: UserSiteSettings) => {
    currentSessionSettings = value;
    translationsToUse = currentSessionSettings.getTranslations().models.webhook;
  },
  adminModeUpdateFunc: (value: boolean) => {
    adminMode = value;
  },
});

let logger = new Logger().withDebugValue(
  'source',
  'src/views/admin/webhooks.svelte',
);

onMount(fetchWebhooks);

// begin experimental API table code

let queryFilter = QueryFilter.fromURLSearchParams();

let apiTableIncrementDisabled: boolean = false;
let apiTableDecrementDisabled: boolean = false;

function searchWebhooks() {
  logger.debug('searchWebhooks called');
}

function incrementPage() {
  if (!apiTableIncrementDisabled) {
    logger.debug(`incrementPage called`);
    queryFilter.page += 1;
    fetchWebhooks();
  }
}

function decrementPage() {
  if (!apiTableDecrementDisabled) {
    logger.debug(`decrementPage called`);
    queryFilter.page -= 1;
    fetchWebhooks();
  }
}

function fetchWebhooks() {
  logger.debug('fetchWebhooks called');

  if (superstore.frontendOnlyMode) {
    webhooks = fakeWebhookFactory.buildList(queryFilter.limit);
  } else {
    V1APIClient.fetchListOfWebhooks(queryFilter, adminMode)
      .then((response: AxiosResponse<WebhookList>) => {
        webhooks = response.data.webhooks || [];

        queryFilter.page = response.data.page;
        apiTableIncrementDisabled = webhooks.length === 0;
        apiTableDecrementDisabled = queryFilter.page === 1;
      })
      .catch((error: AxiosError) => {
        webhookRetrievalError = error.response?.data;
      });
  }
}

function promptDelete(id: number) {
  logger.debug('promptDelete called');

  if (confirm(`are you sure you want to delete webhook #${id}?`)) {
    if (superstore.frontendOnlyMode) {
      fetchWebhooks();
    } else {
      V1APIClient.deleteWebhook(id)
        .then((response: AxiosResponse<Webhook>) => {
          if (response?.status === statusCodes.NO_CONTENT) {
            fetchWebhooks();
          }
        })
        .catch((error: AxiosError<ErrorResponse>) => {
          webhookRetrievalError = error.response?.data?.message || '';
        });
    }
  }
}
</script>

<div class="flex flex-wrap mt-4">
  <div class="w-full mb-12 px-4">
    <APITable
      title="Webhooks"
      headers="{Webhook.headers(translationsToUse)}"
      rows="{webhooks}"
      creationLink="{frontendRoutes.CREATE_WEBHOOK}"
      individualPageLink="/account/webhooks"
      dataRetrievalError="{webhookRetrievalError}"
      searchFunction="{searchWebhooks}"
      incrementDisabled="{apiTableIncrementDisabled}"
      decrementDisabled="{apiTableDecrementDisabled}"
      incrementPageFunction="{incrementPage}"
      decrementPageFunction="{decrementPage}"
      fetchFunction="{fetchWebhooks}"
      deleteFunction="{promptDelete}"
      rowRenderFunction="{Webhook.asRow}"
    />
  </div>
</div>
