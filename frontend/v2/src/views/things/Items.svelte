<script lang="typescript">
    // core components
    import axios, {AxiosError, AxiosResponse} from "axios";
    import {onMount} from "svelte";

    import ItemsTable from "../../components/Tables/ItemsTable.svelte";

    import {inheritQueryFilterSearchParams} from "../../utils";
    import {Item, ItemList, fakeItemFactory } from "../../models";

    export let location;

    let adminMode: boolean = false;
    let itemRetrievalError = '';
    let items: Item[] = []; // fakeItemFactory.buildList(10);

    import { authStatus } from "../../stores";
    let currentAuthStatus = {};
    authStatus.subscribe((value: AuthStatus) => {
        currentAuthStatus = value;
    });

    onMount(() => {
        console.debug("views/things/items.onMount called");

        const path: string = "/api/v1/items";

        const pageURLParams: URLSearchParams = new URLSearchParams(window.location.search);
        const outboundURLParams: URLSearchParams = inheritQueryFilterSearchParams(pageURLParams);

        if (adminMode) {
            outboundURLParams.set("admin", "true");
        }

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
</script>

<div class="flex flex-wrap mt-4">
    <div class="w-full mb-12 px-4">
        <ItemsTable items={items} adminMode={adminMode}/>
    </div>
</div>