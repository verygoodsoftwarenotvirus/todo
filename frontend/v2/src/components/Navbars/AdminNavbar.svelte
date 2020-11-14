<script lang="typescript">
  import { link } from 'svelte-routing';
  import { onDestroy } from 'svelte';

  // core components
  import UserDropdown from '../Dropdowns/UserDropdown.svelte';

  import { sessionSettingsStore, userStatusStore } from '../../stores';
  import { UserSiteSettings, UserStatus } from '../../types';
  import { translations } from '../../i18n';

  let currentAuthStatus = {};
  const unsubscribeFromUserStatusUpdates = userStatusStore.subscribe(
    (value: UserStatus) => {
      currentAuthStatus = value;
    },
  );
  onDestroy(unsubscribeFromUserStatusUpdates);

  // set up translations
  let currentSessionSettings = new UserSiteSettings();
  let translationsToUse = translations.messagesFor(
    currentSessionSettings.language,
  ).components.navbars.adminNavbar;
  const unsubscribeFromSettingsUpdates = sessionSettingsStore.subscribe(
    (value: UserSiteSettings) => {
      currentSessionSettings = value;
      translationsToUse = translations.messagesFor(
        currentSessionSettings.language,
      ).components.navbars.adminNavbar;
    },
  );
  onDestroy(unsubscribeFromSettingsUpdates);
</script>

<!-- Navbar -->
<nav
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
      {translationsToUse.dashboard}
    </a>
    {/if}
    <!-- User -->
    <ul class="flex-col md:flex-row list-none items-center hidden md:flex">
      <UserDropdown />
    </ul>
  </div>
</nav>
<div class="relative bg-red-500 md:pt-32 pb-32 pt-12">
  <div class="px-4 md:px-10 mx-auto w-full" />
</div>
<!-- End Navbar -->
