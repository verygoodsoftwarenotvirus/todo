<script lang="typescript">
// core components
import { AxiosError, AxiosResponse } from 'axios';
import { onDestroy, onMount } from 'svelte';

import {
  ErrorResponse,
  User,
  UserList,
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
import { V1APIClient } from '../../apiClient';
import { translations } from '../../i18n';

import APITable from '../../components/APITable/APITable.svelte';
import { statusCodes } from '../../constants';

export let location;

let userRetrievalError = '';
let users: User[] = [];

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
).models.user;
const unsubscribeFromSettingsUpdates = sessionSettingsStore.subscribe(
  (value: UserSiteSettings) => {
    currentSessionSettings = value;
    translationsToUse = translations.messagesFor(
      currentSessionSettings.language,
    ).models.user;
  },
);
// onDestroy(unsubscribeFromSettingsUpdates);

let logger = new Logger().withDebugValue(
  'source',
  'src/views/admin/Users.svelte',
);

// begin experimental API table code

let queryFilter = QueryFilter.fromURLSearchParams();

let apiTableIncrementDisabled: boolean = false;
let apiTableDecrementDisabled: boolean = false;
let apiTableSearchQuery: string = '';

function searchUsers() {
  logger.debug('searchUsers called');
  //   V1APIClient.searchForUsers(apiTableSearchQuery, queryFilter, adminMode)
  //     .then((response: AxiosResponse<UserList>) => {
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
    fetchUsers();
  }
}

function decrementPage() {
  if (!apiTableDecrementDisabled) {
    logger.debug(`decrementPage called`);
    queryFilter.page -= 1;
    fetchUsers();
  }
}

function fetchUsers() {
  logger.debug('fetchUsers called');

  V1APIClient.fetchListOfUsers(queryFilter, adminMode)
    .then((response: AxiosResponse<UserList>) => {
      users = response.data.users || [];

      queryFilter.page = response.data.page;
      apiTableIncrementDisabled = users.length === 0;
      apiTableDecrementDisabled = queryFilter.page === 1;
    })
    .catch((error: AxiosError) => {
      if (error.response && error.response.data) {
        userRetrievalError = error.response.data;
      }
    });
}

function promptDelete(id: number) {
  logger.debug('promptDelete called');

  if (confirm(`are you sure you want to delete user #${id}?`)) {
    V1APIClient.deleteUser(id)
      .then((response: AxiosResponse<User>) => {
        if (response.status === statusCodes.NO_CONTENT) {
          fetchUsers();
        }
      })
      .catch((error: AxiosError<ErrorResponse>) => {
        if (error.response) {
          if (error.response.data) {
            userRetrievalError = error.response.data.message;
          }
        }
      });
  }
}
</script>

<div class="flex flex-wrap mt-4">
  <div class="w-full mb-12 px-4">
    <APITable
      title="Users"
      headers="{User.headers(translationsToUse)}"
      rows="{users}"
      individualPageLink="/admin/users"
      newPageLink="/admin/users/new"
      dataRetrievalError="{userRetrievalError}"
      searchFunction="{searchUsers}"
      incrementDisabled="{apiTableIncrementDisabled}"
      decrementDisabled="{apiTableDecrementDisabled}"
      incrementPageFunction="{incrementPage}"
      decrementPageFunction="{decrementPage}"
      fetchFunction="{fetchUsers}"
      deleteFunction="{promptDelete}"
      rowRenderFunction="{User.asRow}"
    />
  </div>
</div>
