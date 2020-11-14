<script lang="typescript">
  // core components
  import { AxiosError, AxiosResponse } from 'axios';
  import { onDestroy, onMount } from 'svelte';

  import {
    ErrorResponse,
    AuditLogEntry,
    AuditLogEntryList,
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

  let userRetrievalError = '';
  let users: AuditLogEntry[] = [];

  const useAPITable: boolean = true;

  let currentAuthStatus = {};
  const unsubscribeFromUserStatusUpdates = userStatusStore.subscribe(
    (value: UserStatus) => {
      currentAuthStatus = value;
    },
  );
  onDestroy(unsubscribeFromUserStatusUpdates);

  let adminMode = false;
  const unsubscribeFromAdminModeUpdates = adminModeStore.subscribe(
    (value: boolean) => {
      adminMode = value;
    },
  );
  onDestroy(unsubscribeFromAdminModeUpdates);

  // set up translations
  let currentSessionSettings = new UserSiteSettings();
  let translationsToUse = translations.messagesFor(
    currentSessionSettings.language,
  ).models.user;
  const unsubscribeFromSettingsUpdates = sessionSettingsStore.subscribe(
    (value: UserSiteSettings) => {
      currentSessionSettings = value;
      translationsToUse = translations.messagesFor(
        currentSessionSettings.language,
      ).models.user;
    },
  );
  onDestroy(unsubscribeFromSettingsUpdates);

  let logger = new Logger().withDebugValue(
    'source',
    'src/views/things/AuditLogEntrys.svelte',
  );

  onMount(fetchAuditLogEntrys);

  // begin experimental API table code

  let queryFilter = QueryFilter.fromURLSearchParams();

  let apiTableIncrementDisabled: boolean = false;
  let apiTableDecrementDisabled: boolean = false;
  let apiTableSearchQuery: string = '';

  function searchAuditLogEntrys() {
    logger.debug('searchAuditLogEntrys called');
    //
    //   V1APIClient.searchForAuditLogEntrys(apiTableSearchQuery, queryFilter, adminMode)
    //     .then((response: AxiosResponse<AuditLogEntryList>) => {
    //       users = response.data.users || [];
    //       queryFilter.page = -1;
    //     })
    //     .catch((error: AxiosError) => {
    //       if (error.response) {
    //         if (error.response.data) {
    //           userRetrievalError = error.response.data;
    //         }
    //       }
    //     });
  }

  function incrementPage() {
    if (!apiTableIncrementDisabled) {
      logger.debug(`incrementPage called`);
      queryFilter.page += 1;
      fetchAuditLogEntrys();
    }
  }

  function decrementPage() {
    if (!apiTableDecrementDisabled) {
      logger.debug(`decrementPage called`);
      queryFilter.page -= 1;
      fetchAuditLogEntrys();
    }
  }

  function fetchAuditLogEntrys() {
    logger.debug('fetchAuditLogEntrys called');

    V1APIClient.fetchListOfAuditLogEntries(queryFilter, adminMode)
      .then((response: AxiosResponse<AuditLogEntryList>) => {
        users = response.data.users || [];

        queryFilter.page = response.data.page;
        apiTableIncrementDisabled = users.length === 0;
        apiTableDecrementDisabled = queryFilter.page === 1;
      })
      .catch((error: AxiosError) => {
        if (error.response) {
          if (error.response.data) {
            userRetrievalError = error.response.data;
          }
        }
      });
  }

  function promptDelete(id: number) {
    logger.debug('promptDelete called');
  }
</script>

<div class="flex flex-wrap mt-4">
  <div class="w-full mb-12 px-4">
    <APITable
      title="AuditLogEntrys"
      headers={AuditLogEntry.headers(translationsToUse)}
      rows={users}
      individualPageLink="/things/users"
      newPageLink="/things/users/new"
      dataRetrievalError={userRetrievalError}
      searchFunction={searchAuditLogEntrys}
      incrementDisabled={apiTableIncrementDisabled}
      decrementDisabled={apiTableDecrementDisabled}
      incrementPageFunction={incrementPage}
      decrementPageFunction={decrementPage}
      fetchFunction={fetchAuditLogEntrys}
      deleteFunction={promptDelete}
      rowRenderFunction={AuditLogEntry.asRow} />
  </div>
</div>
