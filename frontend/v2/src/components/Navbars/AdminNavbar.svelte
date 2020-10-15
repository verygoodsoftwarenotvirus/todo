<script lang="typescript">
  import { link } from "svelte-routing";
  import { onDestroy } from "svelte";

  // core components
  import UserDropdown from "../Dropdowns/UserDropdown.svelte";

  import { authStatusStore } from "../../stores";
  import {AuthStatus} from "../../models";
  let currentAuthStatus = {};
  const unsubscribeFromAuthStatusUpdates = authStatusStore.subscribe((value: AuthStatus) => {
    currentAuthStatus = value;
  });
  // onDestroy(unsubscribeFromAuthStatusUpdates);
</script>

<!-- Navbar -->
<navre
  class="absolute top-0 left-0 w-full z-10 bg-transparent md:flex-row md:flex-no-wrap md:justify-start flex items-center p-4"
>
  <div
    class="w-full mx-autp items-center flex justify-between md:flex-no-wrap flex-wrap md:px-10 px-4"
  >
    <!-- Brand -->
    {#if currentAuthStatus.isAdmin}
    <a
      class="text-white text-sm uppercase hidden lg:inline-block font-semibold"
      use:link
      href="/admin"
    >
      Dashboard
    </a>
    {/if}
    <!-- User -->
    <ul class="flex-col md:flex-row list-none items-center hidden md:flex">
      <UserDropdown />
    </ul>
  </div>
</navre>
<!-- End Navbar -->
