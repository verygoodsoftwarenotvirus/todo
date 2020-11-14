<script lang="typescript">
  // core components
  import { AxiosError, AxiosResponse } from 'axios';
  import { onDestroy, onMount } from 'svelte';

  import {
    ErrorResponse,
    OAuth2Client,
    OAuth2ClientList,
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

  let oauth2ClientRetrievalError = '';
  let oauth2Clients: OAuth2Client[] = [];

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
  ).models.oauth2Client;
  const unsubscribeFromSettingsUpdates = sessionSettingsStore.subscribe(
    (value: UserSiteSettings) => {
      currentSessionSettings = value;
      translationsToUse = translations.messagesFor(
        currentSessionSettings.language,
      ).models.oauth2Client;
    },
  );
  //  onDestroy(unsubscribeFromSettingsUpdates);

  let logger = new Logger().withDebugValue(
    'source',
    'src/views/things/OAuth2Clients.svelte',
  );

  onMount(fetchOAuth2Clients);

  // begin experimental API table code

  let queryFilter = QueryFilter.fromURLSearchParams();

  let apiTableIncrementDisabled: boolean = false;
  let apiTableDecrementDisabled: boolean = false;
  let apiTableSearchQuery: string = '';

  function searchOAuth2Clients() {
    logger.debug('searchOAuth2Clients called');
    //
    //   V1APIClient.searchForOAuth2Clients(apiTableSearchQuery, queryFilter, adminMode)
    //     .then((response: AxiosResponse<OAuth2ClientList>) => {
    //       oauth2Clients = response.data.oauth2Clients || [];
    //       queryFilter.page = -1;
    //     })
    //     .catch((error: AxiosError) => {
    //       if (error.response) {
    //         if (error.response.data) {
    //           oauth2ClientRetrievalError = error.response.data;
    //         }
    //       }
    //     });
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

    V1APIClient.fetchListOfOAuth2Clients(queryFilter, adminMode)
      .then((response: AxiosResponse<OAuth2ClientList>) => {
        oauth2Clients = response.data.clients || [];

        queryFilter.page = response.data.page;
        apiTableIncrementDisabled = oauth2Clients.length === 0;
        apiTableDecrementDisabled = queryFilter.page === 1;
      })
      .catch((error: AxiosError) => {
        if (error.response) {
          if (error.response.data) {
            oauth2ClientRetrievalError = error.response.data;
          }
        }
      });
  }

  function promptDelete(id: number) {
    logger.debug('promptDelete called');

    if (confirm(`are you sure you want to delete oauth2Client #${id}?`)) {
      const path: string = `/api/v1/oauth2Clients/${id}`;

      V1APIClient.deleteOAuth2Client(id)
        .then((response: AxiosResponse<OAuth2Client>) => {
          if (response.status === 204) {
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
      headers={OAuth2Client.headers(translationsToUse)}
      rows={oauth2Clients}
      individualPageLink="/things/oauth2Clients"
      newPageLink="/things/oauth2Clients/new"
      dataRetrievalError={oauth2ClientRetrievalError}
      searchFunction={searchOAuth2Clients}
      incrementDisabled={apiTableIncrementDisabled}
      decrementDisabled={apiTableDecrementDisabled}
      incrementPageFunction={incrementPage}
      decrementPageFunction={decrementPage}
      fetchFunction={fetchOAuth2Clients}
      deleteFunction={promptDelete}
      rowRenderFunction={OAuth2Client.asRow} />
  </div>
</div>
