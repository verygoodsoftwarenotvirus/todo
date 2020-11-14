<script lang="typescript">
  // core components
  import { AxiosError, AxiosResponse } from "axios";
  import { onDestroy, onMount } from "svelte";

  import {
    ErrorResponse,
    Item,
    ItemList,
    QueryFilter,
    UserSiteSettings,
    UserStatus,
  } from "../../types";
  import {
    adminModeStore,
    sessionSettingsStore,
    userStatusStore,
  } from "../../stores";
  import { Logger } from "../../logger";
  import { V1APIClient } from "../../requests";
  import { translations } from "../../i18n";

  import APITable from "../../components/APITable/APITable.svelte";

  export let location;

  let itemRetrievalError = "";
  let items: Item[] = [];

  const useAPITable: boolean = true;

  let currentAuthStatus = {};
  const unsubscribeFromUserStatusUpdates = userStatusStore.subscribe(
    (value: UserStatus) => {
      currentAuthStatus = value;
    }
  );
  onDestroy(unsubscribeFromUserStatusUpdates);

  let adminMode = false;
  const unsubscribeFromAdminModeUpdates = adminModeStore.subscribe(
    (value: boolean) => {
      adminMode = value;
    }
  );
  onDestroy(unsubscribeFromAdminModeUpdates);

  // set up translations
  let currentSessionSettings = new UserSiteSettings();
  let translationsToUse = translations.messagesFor(
    currentSessionSettings.language
  ).models.item;
  const unsubscribeFromSettingsUpdates = sessionSettingsStore.subscribe(
    (value: UserSiteSettings) => {
      currentSessionSettings = value;
      translationsToUse = translations.messagesFor(
        currentSessionSettings.language
      ).models.item;
    }
  );
  onDestroy(unsubscribeFromSettingsUpdates);

  let logger = new Logger().withDebugValue(
    "source",
    "src/views/things/Items.svelte"
  );

  onMount(fetchItems);

  // begin experimental API table code

  let queryFilter = QueryFilter.fromURLSearchParams();

  let apiTableIncrementDisabled: boolean = false;
  let apiTableDecrementDisabled: boolean = false;
  let apiTableSearchQuery: string = "";

  function searchItems() {
    logger.debug("searchItems called");

    V1APIClient.searchForItems(apiTableSearchQuery, queryFilter, adminMode)
      .then((response: AxiosResponse<ItemList>) => {
        items = response.data.items || [];
        queryFilter.page = -1;
      })
      .catch((error: AxiosError) => {
        if (error.response) {
          if (error.response.data) {
            itemRetrievalError = error.response.data;
          }
        }
      });
  }

  function incrementPage() {
    if (!apiTableIncrementDisabled) {
      logger.debug(`incrementPage called`);
      queryFilter.page += 1;
      fetchItems();
    }
  }

  function decrementPage() {
    if (!apiTableDecrementDisabled) {
      logger.debug(`decrementPage called`);
      queryFilter.page -= 1;
      fetchItems();
    }
  }

  function fetchItems() {
    logger.debug("fetchItems called");

    V1APIClient.fetchListOfItems(queryFilter, adminMode)
      .then((response: AxiosResponse<ItemList>) => {
        items = response.data.items || [];

        queryFilter.page = response.data.page;
        apiTableIncrementDisabled = items.length === 0;
        apiTableDecrementDisabled = queryFilter.page === 1;
      })
      .catch((error: AxiosError) => {
        if (error.response) {
          if (error.response.data) {
            itemRetrievalError = error.response.data;
          }
        }
      });
  }

  function promptDelete(id: number) {
    logger.debug("promptDelete called");

    if (confirm(`are you sure you want to delete item #${id}?`)) {
      const path: string = `/api/v1/items/${id}`;

      V1APIClient.deleteItem(id)
        .then((response: AxiosResponse<Item>) => {
          if (response.status === 204) {
            fetchItems();
          }
        })
        .catch((error: AxiosError<ErrorResponse>) => {
          if (error.response) {
            if (error.response.data) {
              itemRetrievalError = error.response.data.message;
            }
          }
        });
    }
  }
</script>

<div class="flex flex-wrap mt-4">
  <div class="w-full mb-12 px-4">
    <APITable
      title="Items"
      headers={Item.headers(translationsToUse)}
      rows={items}
      individualPageLink="/things/items"
      newPageLink="/things/items/new"
      dataRetrievalError={itemRetrievalError}
      searchFunction={searchItems}
      incrementDisabled={apiTableIncrementDisabled}
      decrementDisabled={apiTableDecrementDisabled}
      incrementPageFunction={incrementPage}
      decrementPageFunction={decrementPage}
      fetchFunction={fetchItems}
      deleteFunction={promptDelete}
      rowRenderFunction={Item.asRow} />
  </div>
</div>
