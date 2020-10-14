<script lang="typescript">
  import { navigate } from "svelte-routing";
  import { onMount } from "svelte";
  import axios, {AxiosError, AxiosResponse} from "axios";

  import { Item } from "../../../models";

  export let id: number = 0;

  // local state
  let item: Item = new Item();
  let apiError: string = '';

  function createItem(): void {
    console.debug(`vies/things/Item.createItem called`);

    const path: string = `/api/v1/items`;

    console.dir(item);

    axios.post(path, item, { withCredentials: true })
            .then((response: AxiosResponse<Item>) => {
              const newItem = response.data;
              navigate( `/things/items/${newItem.id}`, { state: {}, replace: true} );
            })
            .catch((error: AxiosError) => {
              if (error.response) {
                if (error.response.data) {
                  apiError = error.response.data;
                }
              }
            });
  }
</script>

<div class="relative flex flex-col min-w-0 break-words bg-white w-full mb-6 shadow-lg rounded">
  <div class="rounded-t mb-0 px-4 py-3 bg-transparent justify-between ">
    <div class="flex flex-wrap items-center">
      <div class="relative w-full max-w-full flex-grow flex-1">
       <h2 class="text-gray-800 text-xl font-semibold">
          Create Item
        </h2>
      </div>
      <div class="flex w-full max-w-full flex-grow justify-end flex-1">
        <button class="bg-green-500 hover:bg-green-700 text-white font-bold py-2 px-4 rounded" on:click={createItem}><i class="fa fa-plus-circle"></i> Create</button>
      </div>
    </div>
  </div>
  <div>
    <div class="flex flex-wrap -mx-3 mb-6">
      <div class="w-full md:w-1/2 px-3 mb-6 md:mb-0">
        <label class="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2" for="grid-first-name">
          Name
        </label>
        <input class="appearance-none block w-full text-gray-700 border border-red-500 rounded py-3 px-4 mb-3 leading-tight focus:outline-none focus:bg-white" id="grid-first-name" type="text" bind:value={item.name}>
        <!--  <p class="text-red-500 text-xs italic">Please fill out this field.</p>-->
      </div>
      <div class="w-full md:w-1/2 px-3">
        <label class="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2" for="grid-last-name">
          Details
        </label>
        <input class="appearance-none block w-full text-gray-700 border border-gray-200 rounded py-3 px-4 leading-tight focus:outline-none focus:bg-white focus:border-gray-500" id="grid-last-name" type="text" bind:value={item.details}>
      </div>
    </div>
  </div>
</div>
