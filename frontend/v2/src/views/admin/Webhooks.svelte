<script lang="typescript">
  // core components
  import { AxiosError, AxiosResponse } from 'axios';
  import { onDestroy, onMount } from 'svelte';

  import {
    ErrorResponse,
    Webhook,
    WebhookList,
    QueryFilter,
    UserSiteSettings,
    UserStatus,
  } from '../../types';
  import {
    adminModeStore,
    sessionSettingsStore,
    userStatusStore,
  } from '../../stores';
  import { Logger } from '../../logger';
  import { V1APIClient } from '../../requests';
  import { translations } from '../../i18n';

  import APITable from '../../components/APITable/APITable.svelte';

  export let location;

  let webhookRetrievalError = '';
  let webhooks: Webhook[] = [];

  const useAPITable: boolean = true;

  let currentAuthStatus = {};
  const unsubscribeFromUserStatusUpdates = userStatusStore.subscribe(
    (value: UserStatus) => {
      currentAuthStatus = value;
    },
  );
  //  onDestroy(unsubscribeFromUserStatusUpdates);

  let adminMode = false;
  const unsubscribeFromAdminModeUpdates = adminModeStore.subscribe(
    (value: boolean) => {
      adminMode = value;
    },
  );
  //  onDestroy(unsubscribeFromAdminModeUpdates);

  // set up translations
  let currentSessionSettings = new UserSiteSettings();
  let translationsToUse = translations.messagesFor(
    currentSessionSettings.language,
  ).models.webhook;
  const unsubscribeFromSettingsUpdates = sessionSettingsStore.subscribe(
    (value: UserSiteSettings) => {
      currentSessionSettings = value;
      translationsToUse = translations.messagesFor(
        currentSessionSettings.language,
      ).models.webhook;
    },
  );
  //  onDestroy(unsubscribeFromSettingsUpdates);

  let logger = new Logger().withDebugValue(
    'source',
    'src/views/things/Webhooks.svelte',
  );

  onMount(fetchWebhooks);

  // begin experimental API table code

  let queryFilter = QueryFilter.fromURLSearchParams();

  let apiTableIncrementDisabled: boolean = false;
  let apiTableDecrementDisabled: boolean = false;
  let apiTableSearchQuery: string = '';

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

    V1APIClient.fetchListOfWebhooks(queryFilter, adminMode)
      .then((response: AxiosResponse<WebhookList>) => {
        webhooks = response.data.webhooks || [];

        queryFilter.page = response.data.page;
        apiTableIncrementDisabled = webhooks.length === 0;
        apiTableDecrementDisabled = queryFilter.page === 1;
      })
      .catch((error: AxiosError) => {
        if (error.response) {
          if (error.response.data) {
            webhookRetrievalError = error.response.data;
          }
        }
      });
  }

  function promptDelete(id: number) {
    logger.debug('promptDelete called');

    if (confirm(`are you sure you want to delete webhook #${id}?`)) {
      const path: string = `/api/v1/webhooks/${id}`;

      V1APIClient.deleteWebhook(id)
        .then((response: AxiosResponse<Webhook>) => {
          if (response.status === 204) {
            fetchWebhooks();
          }
        })
        .catch((error: AxiosError<ErrorResponse>) => {
          if (error.response) {
            if (error.response.data) {
              webhookRetrievalError = error.response.data.message;
            }
          }
        });
    }
  }
</script>

<div class="flex flex-wrap mt-4">
  <div class="w-full mb-12 px-4">
    <APITable
      title="Webhooks"
      headers={Webhook.headers(translationsToUse)}
      rows={webhooks}
      individualPageLink="/things/webhooks"
      newPageLink="/things/webhooks/new"
      dataRetrievalError={webhookRetrievalError}
      searchFunction={searchWebhooks}
      incrementDisabled={apiTableIncrementDisabled}
      decrementDisabled={apiTableDecrementDisabled}
      incrementPageFunction={incrementPage}
      decrementPageFunction={decrementPage}
      fetchFunction={fetchWebhooks}
      deleteFunction={promptDelete}
      rowRenderFunction={Webhook.asRow} />
  </div>
</div>
