<script lang="typescript">
// core components
import type { AxiosError, AxiosResponse } from 'axios';
import { onMount } from 'svelte';

import {
  ErrorResponse,
  Account,
  AccountList,
  QueryFilter,
  UserSiteSettings,
  UserStatus,
  fakeUserFactory,
  fakeAccountFactory,
} from '../../types';
import { Logger } from '../../logger';
import { V1APIClient } from '../../apiClient';

import APITable from '../core/apiTable/apiTable.svelte';
import {frontendRoutes, statusCodes} from '../../constants';
import { Superstore } from '../../stores';

export let location: Location;

let accountRetrievalError = '';
let accounts: Account[] = [];

let adminMode: boolean = false;
let currentAuthStatus: UserStatus = new UserStatus();
let currentSessionSettings = new UserSiteSettings();
let translationsToUse = currentSessionSettings.getTranslations().models.account;

let superstore = new Superstore({
  userStatusStoreUpdateFunc: (value: UserStatus) => {
    currentAuthStatus = value;
  },
  sessionSettingsStoreUpdateFunc: (value: UserSiteSettings) => {
    currentSessionSettings = value;
    translationsToUse = currentSessionSettings.getTranslations().models.account;
  },
  adminModeUpdateFunc: (value: boolean) => {
    adminMode = value;
  },
});

let logger = new Logger().withDebugValue(
  'source',
  'src/views/admin/accounts.svelte',
);

onMount(fetchAccounts);

// begin experimental API table code

let queryFilter = QueryFilter.fromURLSearchParams();

let apiTableIncrementDisabled: boolean = false;
let apiTableDecrementDisabled: boolean = false;
let apiTableSearchQuery: string = '';

function searchAccounts() {
  logger.debug('searchAccounts called');
}

function incrementPage() {
  if (!apiTableIncrementDisabled) {
    logger.debug(`incrementPage called`);
    queryFilter.page += 1;
    fetchAccounts();
  }
}

function decrementPage() {
  if (!apiTableDecrementDisabled) {
    logger.debug(`decrementPage called`);
    queryFilter.page -= 1;
    fetchAccounts();
  }
}

function fetchAccounts() {
  logger.debug('fetchAccounts called');

  if (superstore.frontendOnlyMode) {
    accounts = fakeAccountFactory.buildList(queryFilter.limit);
  } else {
    V1APIClient.fetchListOfAccounts(queryFilter, adminMode)
      .then((response: AxiosResponse<AccountList>) => {
        accounts = response.data.accounts || [];

        queryFilter.page = response.data.page;
        apiTableIncrementDisabled = accounts.length === 0;
        apiTableDecrementDisabled = queryFilter.page === 1;
      })
      .catch((error: AxiosError) => {
        accountRetrievalError = error.response?.data;
      });
  }
}

function promptDelete(id: number) {
  logger.debug('promptDelete called');

  if (confirm(`are you sure you want to delete account #${id}?`)) {
    if (superstore.frontendOnlyMode) {
      fetchAccounts();
    } else {
      V1APIClient.deleteAccount(id)
        .then((response: AxiosResponse<Account>) => {
          if (response?.status === statusCodes.NO_CONTENT) {
            fetchAccounts();
          }
        })
        .catch((error: AxiosError<ErrorResponse>) => {
          accountRetrievalError = error.response?.data?.message || '';
        });
    }
  }
}
</script>

<div class="flex flex-wrap mt-4">
  <div class="w-full mb-12 px-4">
    <APITable
      title="Accounts"
      headers="{Account.headers(translationsToUse)}"
      rows="{accounts}"
      individualPageLink="{frontendRoutes.ADMIN_INDIVIDUAL_ACCOUNT}"
      dataRetrievalError="{accountRetrievalError}"
      searchFunction="{searchAccounts}"
      incrementDisabled="{apiTableIncrementDisabled}"
      decrementDisabled="{apiTableDecrementDisabled}"
      incrementPageFunction="{incrementPage}"
      decrementPageFunction="{decrementPage}"
      fetchFunction="{fetchAccounts}"
      deleteFunction="{promptDelete}"
      rowRenderFunction="{Account.asRow}"
    />
  </div>
</div>
