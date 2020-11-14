<script lang="typescript">
  import { link } from 'svelte-routing';
  import { onDestroy } from 'svelte';

  // core components
  // import PagesDropdown from "../components/Dropdowns/PagesDropdown.svelte";

  import { UserSiteSettings } from '../../types';
  import { translations } from '../../i18n';
  import { sessionSettingsStore } from '../../stores';

  let navbarOpen: Boolean = false;

  function setNavbarOpen() {
    navbarOpen = !navbarOpen;
  }

  // set up translations
  let currentSessionSettings = new UserSiteSettings();
  let translationsToUse = translations.messagesFor(
    currentSessionSettings.language,
  ).components.navbars.authNavbar;
  const unsubscribeFromSettingsUpdates = sessionSettingsStore.subscribe(
    (value: UserSiteSettings) => {
      currentSessionSettings = value;
      translationsToUse = translations.messagesFor(
        currentSessionSettings.language,
      ).components.navbars.authNavbar;
    },
  );
  onDestroy(unsubscribeFromSettingsUpdates);
</script>

<nav
  class="top-0 absolute z-50 w-full flex flex-wrap items-center justify-between px-2 py-3 navbar-expand-lg">
  <div
    class="container px-4 mx-auto flex flex-wrap items-center justify-between">
    <div
      class="w-full relative flex justify-between lg:w-auto lg:static lg:block lg:justify-start">
      <a
        use:link
        class="text-white text-sm font-bold leading-relaxed inline-block mr-4 py-2 whitespace-no-wrap uppercase"
        href="/">
        {translationsToUse.serviceName}
      </a>
      <button
        class="cursor-pointer text-xl leading-none px-3 py-1 border border-solid border-transparent rounded bg-transparent block lg:hidden outline-none focus:outline-none"
        type="button"
        on:click={setNavbarOpen}>
        <i class="text-white fas fa-bars" />
      </button>
    </div>
  </div>
</nav>
