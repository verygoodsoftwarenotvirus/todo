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

let entryRetrievalError = '';
let entries: AuditLogEntry[] = [];

const useAPITable: boolean = true;

let currentAuthStatus = {};
const unsubscribeFromUserStatusUpdates = userStatusStore.subscribe(
  (value: UserStatus) => {
    currentAuthStatus = value;
  },
);
// onDestroy(unsubscribeFromUserStatusUpdates);

let adminMode = false;
const unsubscribeFromAdminModeUpdates = adminModeStore.subscribe(
  (value: boolean) => {
    adminMode = value;
  },
);
// onDestroy(unsubscribeFromAdminModeUpdates);

// set up translations
let currentSessionSettings = new UserSiteSettings();
let translationsToUse = translations.messagesFor(
  currentSessionSettings.language,
).models.auditLogEntry;

const unsubscribeFromSettingsUpdates = sessionSettingsStore.subscribe(
  (value: UserSiteSettings) => {
    currentSessionSettings = value;
    translationsToUse = translations.messagesFor(
      currentSessionSettings.language,
    ).models.auditLogEntry;
  },
);
// onDestroy(unsubscribeFromSettingsUpdates);

let logger = new Logger().withDebugValue(
  'source',
  'src/views/admin/AuditLogEntries.svelte',
);

onMount(fetchAuditLogEntries);

// begin experimental API table code

let queryFilter = QueryFilter.fromURLSearchParams();

let apiTableIncrementDisabled: boolean = false;
let apiTableDecrementDisabled: boolean = false;
let apiTableSearchQuery: string = '';

function searchAuditLogEntries() {
  logger.debug('searchAuditLogEntries called');
  //   V1APIClient.searchForAuditLogEntries(apiTableSearchQuery, queryFilter, adminMode)
  //     .then((response: AxiosResponse<AuditLogEntryList>) => {
  //       entries = response.data.entries || [];
  //       queryFilter.page = -1;
  //     })
  //     .catch((error: AxiosError) => {
  //       if (error.response) {
  //         if (error.response.data) {
  //           entryRetrievalError = error.response.data;
  //         }
  //       }
  //     });
}

function incrementPage() {
  if (!apiTableIncrementDisabled) {
    logger.debug(`incrementPage called`);
    queryFilter.page += 1;
    fetchAuditLogEntries();
  }
}

function decrementPage() {
  if (!apiTableDecrementDisabled) {
    logger.debug(`decrementPage called`);
    queryFilter.page -= 1;
    fetchAuditLogEntries();
  }
}

function fetchAuditLogEntries() {
  logger.debug('fetchAuditLogEntries called');

  V1APIClient.fetchListOfAuditLogEntries(queryFilter, adminMode)
    .then((response: AxiosResponse<AuditLogEntryList>) => {
      entries = response.data.entries || [];

      queryFilter.page = response.data.page;
      apiTableIncrementDisabled = entries.length === 0;
      apiTableDecrementDisabled = queryFilter.page === 1;
    })
    .catch((error: AxiosError) => {
      if (error.response) {
        if (error.response.data) {
          entryRetrievalError = error.response.data;
        }
      }
    });
}
</script>

<div class="flex flex-wrap mt-4">
  <div class="w-full mb-12 px-4">
    <APITable
      title="Audit Log"
      headers="{AuditLogEntry.headers(translationsToUse)}"
      rows="{entries}"
      individualPageLink="/admin/audit_log_entries"
      newPageLink=""
      dataRetrievalError="{entryRetrievalError}"
      searchEnabled="{false}"
      incrementDisabled="{apiTableIncrementDisabled}"
      decrementDisabled="{apiTableDecrementDisabled}"
      incrementPageFunction="{incrementPage}"
      decrementPageFunction="{decrementPage}"
      fetchFunction="{fetchAuditLogEntries}"
      deleteEnabled="{false}"
      rowRenderFunction="{AuditLogEntry.asRow}"
    />
  </div>
</div>
