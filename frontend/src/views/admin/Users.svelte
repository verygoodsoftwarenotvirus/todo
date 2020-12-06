<script lang="typescript">
// core components
import { AxiosError, AxiosResponse } from 'axios';

import {
  ErrorResponse,
  User,
  UserList,
  QueryFilter,
  UserSiteSettings,
  UserStatus,
  fakeUserFactory,
} from '../../types';
import { Logger } from '../../logger';
import { V1APIClient } from '../../apiClient';

import APITable from '../../components/APITable/APITable.svelte';
import { statusCodes } from '../../constants';
import { Superstore } from '../../stores';

export let location;

let userRetrievalError = '';
let users: User[] = [];

let adminMode: boolean = false;
let currentAuthStatus: UserStatus = new UserStatus();
let currentSessionSettings = new UserSiteSettings();
let translationsToUse = currentSessionSettings.getTranslations().models.user;

let superstore = new Superstore({
  userStatusStoreUpdateFunc: (value: UserStatus) => {
    currentAuthStatus = value;
  },
  sessionSettingsStoreUpdateFunc: (value: UserSiteSettings) => {
    currentSessionSettings = value;
    translationsToUse = currentSessionSettings.getTranslations().models.user;
  },
  adminModeUpdateFunc: (value: boolean) => {
    adminMode = value;
  },
});

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

  if (superstore.frontendOnlyMode) {
    users = fakeUserFactory.buildList(queryFilter.limit);
  } else {
    V1APIClient.fetchListOfUsers(queryFilter, adminMode)
      .then((response: AxiosResponse<UserList>) => {
        users = response.data.users || [];

        queryFilter.page = response.data.page;
        apiTableIncrementDisabled = users.length === 0;
        apiTableDecrementDisabled = queryFilter.page === 1;
      })
      .catch((error: AxiosError) => {
        userRetrievalError = error.response?.data;
      });
  }
}

function promptDelete(id: number) {
  logger.debug('promptDelete called');

  if (confirm(`are you sure you want to delete user #${id}?`)) {
    if (superstore.frontendOnlyMode) {
      fetchUsers();
    } else {
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
}
</script>

<div class="flex flex-wrap mt-4">
  <div class="w-full mb-12 px-4">
    <APITable
      title="Users"
      headers="{User.headers(translationsToUse)}"
      rows="{users}"
      individualPageLink="/admin/users"
      creationLink="/admin/users/new"
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
