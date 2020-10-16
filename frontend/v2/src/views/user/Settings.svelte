<script lang="typescript">
  import { onMount } from "svelte";

  // core components
  import CardSettings from "../../components/Cards/CardSettings.svelte";
  import CardProfile from "../../components/Cards/CardProfile.svelte";

  export let location: Location;

  import { Logger } from "../../logger";

  let logger = new Logger();

  onMount(() => {
    logger.info("views/user/Settings.svelte onMount");
  })
</script>

<div class="flex flex-wrap">
  <div class="w-full lg:w-8/12 px-4">
    <CardSettings />
  </div>
  <div class="w-full lg:w-4/12 px-4">
    <CardProfile />
  </div>
</div>
