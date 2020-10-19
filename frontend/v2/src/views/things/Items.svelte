<script lang="typescript">
    // core components
    import axios, { AxiosError, AxiosResponse } from "axios";
    import { onDestroy, onMount} from "svelte";

    import APITable from "../../components/APITable/APITable.svelte";

    import { inheritQueryFilterSearchParams } from "../../utils";
    import {Item, ItemList, QueryFilter} from "../../models";

    export let location;

    let itemRetrievalError = '';
    let items: Item[] = []; // fakeItemFactory.buildList(10);

    const useAPITable: boolean = true;

    import { authStatusStore } from "../../stores";
    let currentAuthStatus = {};
    const unsubscribeFromAuthStatusUpdates = authStatusStore.subscribe((value: UserStatus) => {
        currentAuthStatus = value;
    });

    import { adminModeStore } from "../../stores";
    let adminMode = false;
    const unsubscribeFromAdminModeUpdates = adminModeStore.subscribe((value: boolean) => {
        adminMode = value;
    });
    // onDestroy(unsubscribeFromAdminModeUpdates);

    import { Logger } from "../../logger";
    let logger = new Logger().withDebugValue("source", "src/views/things/Items.svelte");

    onMount(() => {
        const path: string = "/api/v1/items";

        const pageURLParams: URLSearchParams = new URLSearchParams(window.location.search);
        const outboundURLParams: URLSearchParams = inheritQueryFilterSearchParams(pageURLParams);

        const qs = outboundURLParams.toString()
        const uri = "/api/v1/items?" + qs;

        axios.get(uri, { withCredentials: true })
            .then((response: AxiosResponse<ItemList>) => {
                items = response.data.items || [];
            })
            .catch((error: AxiosError) => {
                if (error.response) {
                    if (error.response.data) {
                        itemRetrievalError = error.response.data;
                    }
                }
            });
    })

    // begin experimental API table code

    let queryFilter = new QueryFilter();

    let apiTableIncrementDisabled: boolean = false;
    let apiTableDecrementDisabled: boolean = false;

    let apiTableSearchQuery: string = '';

    function searchItems() {
        logger.debug("searchItems called");

        const path: string = "/api/v1/items/search";

        const qf = QueryFilter.fromURLSearchParams();
        const outboundURLParams = qf.toURLSearchParams();

        if (adminMode) {
            outboundURLParams.set("admin", "true");
        }
        outboundURLParams.set("q", apiTableSearchQuery)

        const qs = outboundURLParams.toString()
        const uri = `${path}?${qs}`;

        axios.get(uri, { withCredentials: true })
            .then((response: AxiosResponse<ItemList>) => {
                items = response.data || [];
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

        const path: string = "/api/v1/items";

        const qf = QueryFilter.fromURLSearchParams();
        const outboundURLParams = qf.toURLSearchParams();

        if (adminMode) {
            outboundURLParams.set("admin", "true");
        }

        const qs = outboundURLParams.toString()
        const uri = `${path}?${qs}`;

        axios.get(uri, { withCredentials: true })
            .then((response: AxiosResponse<ItemList>) => {
                items = response.data.items || [];

                apiTableCurrentPage = response.data.page;
                apiTableIncrementDisabled = items.length === 0;
                apiTableDecrementDisabled = apiTableCurrentPage === 1;
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

            axios.delete(path, { withCredentials: true })
                .then((response: AxiosResponse<Item>) => {
                    if (response.status === 204) {
                        fetchItems();
                    }
                })
                .catch((error: AxiosError) => {
                    if (error.response) {
                        if (error.response.data) {
                            itemRetrievalError = error.response.data;
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
            headers={Item.headers()}
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
            rowRenderFunction={Item.asRow}
        />
    </div>
</div>