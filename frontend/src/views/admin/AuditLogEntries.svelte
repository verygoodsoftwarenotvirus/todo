<script lang="typescript">
// core components
import { AxiosError, AxiosResponse } from 'axios';
import { onMount } from 'svelte';

import {
  AuditLogEntry,
  AuditLogEntryList,
  fakeAuditLogEntryFactory,
  QueryFilter,
  UserSiteSettings,
  UserStatus,
} from '../../types';
import { Logger } from '../../logger';
import { V1APIClient } from '../../apiClient';

import APITable from '../../components/APITable/APITable.svelte';
import { Superstore } from '../../stores/superstore';

export let location;

let entryRetrievalError = '';
let entries: AuditLogEntry[] = [];

let adminMode: boolean = false;
let currentAuthStatus: UserStatus = new UserStatus();
let currentSessionSettings = new UserSiteSettings();
let translationsToUse = currentSessionSettings.getTranslations().models
  .auditLogEntry;

let superstore = new Superstore({
  userStatusStoreUpdateFunc: (value: UserStatus) => {
    currentAuthStatus = value;
  },
  sessionSettingsStoreUpdateFunc: (value: UserSiteSettings) => {
    currentSessionSettings = value;
    translationsToUse = currentSessionSettings.getTranslations().models
      .auditLogEntry;
  },
  adminModeUpdateFunc: (value: boolean) => {
    adminMode = value;
  },
});

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

  if (superstore.frontendOnlyMode) {
    entries = fakeAuditLogEntryFactory.buildList(queryFilter.limit);
  } else {
    V1APIClient.fetchListOfAuditLogEntries(queryFilter, adminMode)
      .then((response: AxiosResponse<AuditLogEntryList>) => {
        entries = response.data.entries || [];

        queryFilter.page = response.data.page;
        apiTableIncrementDisabled = entries.length === 0;
        apiTableDecrementDisabled = queryFilter.page === 1;
      })
      .catch((error: AxiosError) => {
        entryRetrievalError = error.response?.data;
      });
  }
}
</script>

<div class="flex flex-wrap mt-4">
  <div class="w-full mb-12 px-4">
    <APITable
      title="Audit Log"
      headers="{AuditLogEntry.headers(translationsToUse)}"
      rows="{entries}"
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
