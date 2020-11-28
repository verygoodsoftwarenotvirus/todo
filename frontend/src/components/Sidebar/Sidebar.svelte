<script lang="typescript">
import { link } from 'svelte-routing';

// core components
import UserDropdown from '../Dropdowns/UserDropdown.svelte';

import { UserSiteSettings, UserStatus } from '../../types';
import { Logger } from '../../logger';
import { Superstore } from '../../stores/superstore';

export let location: Location;

let collapseShow: string = 'hidden';
function toggleCollapseShow(classes) {
  collapseShow = classes;
}

let logger = new Logger().withDebugValue(
  'source',
  'src/components/Sidebar/Sidebar.svelte',
);

let currentAuthStatus = new UserStatus();
let currentSessionSettings = new UserSiteSettings();
let translationsToUse = currentSessionSettings.getTranslations().components
  .sidebars.primary;

let superstore = new Superstore({
  userStatusStoreUpdateFunc: (value: UserStatus) => {
    currentAuthStatus = value;
  },
  sessionSettingsStoreUpdateFunc: (value: UserSiteSettings) => {
    currentSessionSettings = value;
    translationsToUse = currentSessionSettings.getTranslations().components
      .sidebars.primary;
  },
});
</script>

s
<nav
  class="md:left-0 md:block md:fixed md:top-0 md:bottom-0 md:overflow-y-auto md:flex-row md:flex-no-wrap md:overflow-hidden shadow-xl bg-white flex flex-wrap items-center justify-between relative md:w-64 z-10 py-4 px-6"
>
  <div
    class="md:flex-col md:items-stretch md:min-h-full md:flex-no-wrap px-0 flex flex-wrap items-center justify-between w-full mx-auto"
  >
    <!-- Toggler -->
    <button
      class="cursor-pointer text-black opacity-50 md:hidden px-3 py-1 text-xl leading-none bg-transparent rounded border border-solid border-transparent"
      type="button"
      on:click="{() => toggleCollapseShow('bg-white m-2 py-3 px-6')}"
    >
      <i class="fas fa-bars"></i>
    </button>
    <!-- Brand -->
    <a
      use:link
      class="md:block text-left md:pb-2 text-gray-700 mr-0 inline-block whitespace-no-wrap text-sm uppercase font-bold p-4 px-0"
      href="/"
    >
      {translationsToUse.serviceName}
    </a>
    <!-- User -->
    <ul class="md:hidden items-center flex flex-wrap list-none">
      <li class="inline-block relative">
        <!-- <NotificationDropdown /> -->
      </li>
      <li class="inline-block relative">
        <UserDropdown />
      </li>
    </ul>
    <!-- Collapse -->
    <div
      class="md:flex md:flex-col md:items-stretch md:opacity-100 md:relative md:mt-4 md:shadow-none shadow absolute top-0 left-0 right-0 z-40 overflow-y-auto overflow-x-hidden h-auto items-center flex-1 rounded {collapseShow}"
    >
      <div>
        <h6
          class="md:min-w-full text-gray-600 text-xs uppercase font-bold block pt-1 pb-4 no-underline"
        >
          {translationsToUse.things}
        </h6>
        <!-- Navigation -->

        <ul class="md:flex-col md:min-w-full flex flex-col list-none md:mb-4">
          <li class="items-center">
            <a
              use:link
              class="text-gray-800 hover:text-gray-600 text-xs uppercase py-3 font-bold block"
              href="/things/items"
            >
              <i class="fas fa-list-ul text-gray-400 mr-2 text-sm"></i>
              {translationsToUse.items}
            </a>
          </li>
        </ul>
      </div>

      {#if currentAuthStatus.isAdmin && adminMode}
        <hr class="my-4 md:min-w-full" />
        <div>
          <h6
            class="md:min-w-full text-gray-600 text-xs uppercase font-bold block pt-1 pb-4 no-underline"
          >
            {translationsToUse.admin}
          </h6>

          <!-- USERS -->

          <ul class="md:flex-col md:min-w-full flex flex-col list-none md:mb-4">
            <li class="items-center">
              <a
                use:link
                class="text-gray-800 hover:text-gray-600 text-xs uppercase py-3 font-bold block"
                href="/admin/users"
              >
                <i class="fas fa-users text-gray-400 mr-2 text-sm"></i>
                {translationsToUse.users}
              </a>
            </li>
          </ul>

          <!-- OAUTH2 CLIENTS -->

          <ul class="md:flex-col md:min-w-full flex flex-col list-none md:mb-4">
            <li class="items-center">
              <a
                use:link
                class="text-gray-800 hover:text-gray-600 text-xs uppercase py-3 font-bold block"
                href="/admin/oauth2_clients"
              >
                <i class="fas fa-robot text-gray-400 mr-2 text-sm"></i>
                {translationsToUse.oauth2Clients}
              </a>
            </li>
          </ul>

          <!-- WEBHOOKS -->

          <ul class="md:flex-col md:min-w-full flex flex-col list-none md:mb-4">
            <li class="items-center">
              <a
                use:link
                class="text-gray-800 hover:text-gray-600 text-xs uppercase py-3 font-bold block"
                href="/admin/webhooks"
              >
                <i class="fas fa-network-wired text-gray-400 mr-2 text-sm"></i>
                {translationsToUse.webhooks}
              </a>
            </li>
          </ul>

          <!-- AUDIT LOG -->

          <ul class="md:flex-col md:min-w-full flex flex-col list-none md:mb-4">
            <li class="items-center">
              <a
                use:link
                class="text-gray-800 hover:text-gray-600 text-xs uppercase py-3 font-bold block"
                href="/admin/audit_log"
              >
                <i class="fas fa-record-vinyl text-gray-400 mr-2 text-sm"></i>
                {translationsToUse.auditLog}
              </a>
            </li>
          </ul>

          <!-- SERVER SETTINGS -->

          <ul class="md:flex-col md:min-w-full flex flex-col list-none md:mb-4">
            <li class="items-center">
              <a
                use:link
                class="text-gray-800 hover:text-gray-600 text-xs uppercase py-3 font-bold block"
                href="/admin/settings"
              >
                <i class="fas fa-cog text-gray-400 mr-2 text-sm"></i>
                {translationsToUse.serverSettings}
              </a>
            </li>
          </ul>
        </div>
      {/if}
    </div>
  </div>
</nav>
