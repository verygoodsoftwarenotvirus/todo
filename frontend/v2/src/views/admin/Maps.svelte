<script lang="typescript">
  // core components
  import MapExample from "components/Maps/MapExample.svelte";

  export let location: Location;
</script>

<div class="flex flex-wrap">
  <div class="w-full px-4">
    <div
      class="relative flex flex-col min-w-0 break-words bg-white w-full mb-6 shadow-lg rounded"
    >
      <MapExample />
    </div>
  </div>
</div>
