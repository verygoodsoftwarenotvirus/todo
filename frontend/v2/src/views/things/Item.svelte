<script lang="typescript">
  import { onMount } from "svelte";
  import axios, {AxiosError, AxiosResponse} from "axios";

  import { Item } from "../../models";

  export let id: number = 0;
  let itemRetrievalError: string = '';
  let item: Item = new Item();

  onMount(async() => {
    console.debug(`components/cards/Item onMount called for item #${id}`);

    if (id === 0) {
      throw new Error("id cannot be zero!");
    }

    const path: string = `/api/v1/items/${id}`;

    axios.get(path, { withCredentials: true })
            .then((response: AxiosResponse<Item>) => {
              item = response.data;
              console.dir(item);
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

<div
  class="relative flex flex-col min-w-0 break-words bg-white w-full mb-6 shadow-lg rounded"
>
  <div class="rounded-t mb-0 px-4 py-3 bg-transparent">
    <div class="flex flex-wrap items-center">
      <div class="relative w-full max-w-full flex-grow flex-1">
        <h2 class="text-gray-800 text-xl font-semibold">
          #{item.id}: { item.name }
        </h2>
      </div>
    </div>
  </div>
  <div class="p-4 flex-auto">
    <div class="relative h-350-px">
      <canvas id="bar-chart"></canvas>
    </div>
  </div>
</div>
