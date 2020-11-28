<script lang="typescript">
import type { AxiosResponse } from 'axios';
import ClickOutside from 'svelte-click-outside';
import { navigate } from 'svelte-routing';
import { createPopper } from '@popperjs/core'; // library for creating dropdown menu appear on click

import { V1APIClient } from '../../apiClient';
import { UserSiteSettings, UserStatus } from '../../types';

let dropdownPopoverShow: Boolean = false;

let triggerEl: HTMLElement;
let btnDropdownRef: HTMLElement;
let popoverDropdownRef: HTMLElement;

import { Logger } from '../../logger';
import { frontendRoutes, statusCodes } from '../../constants';
import { Superstore } from '../../stores/superstore';
let logger = new Logger().withDebugValue(
  'source',
  'src/components/Dropdowns/UserDropdown.svelte',
);

let adminMode: boolean = false;
let currentAuthStatus: UserStatus = new UserStatus();
let currentSessionSettings = new UserSiteSettings();
let translationsToUse = currentSessionSettings.getTranslations().components
  .dropdowns.userDropdown;

let superstore = new Superstore({
  userStatusStoreUpdateFunc: (value: UserStatus) => {
    currentAuthStatus = value;
  },
  sessionSettingsStoreUpdateFunc: (value: UserSiteSettings) => {
    currentSessionSettings = value;
    translationsToUse = currentSessionSettings.getTranslations().components
      .dropdowns.userDropdown;
  },
  adminModeUpdateFunc: (value: boolean) => {
    adminMode = value;
  },
});

function goToSettings() {
  dropdownPopoverShow = false;
  logger.debug(
    `navigating to ${frontendRoutes.USER_SETTINGS} via goToSettings`,
  );
  navigate(frontendRoutes.USER_SETTINGS, { state: {}, replace: true });
}

function logout() {
  V1APIClient.logout().then((response: AxiosResponse) => {
    if (response.status === statusCodes.OK) {
      logger.debug(
        `navigating to ${frontendRoutes.LOGIN} via logout promise resolution`,
      );
      navigate(frontendRoutes.LOGIN, { state: {}, replace: true });
      dropdownPopoverShow = false;
    }
  });
}

const hideDropdown = () => {
  dropdownPopoverShow = false;
};

const toggleDropdown = (event: any) => {
  event.preventDefault();
  if (dropdownPopoverShow) {
    dropdownPopoverShow = false;
  } else {
    dropdownPopoverShow = true;
    createPopper(btnDropdownRef, popoverDropdownRef, {
      placement: 'bottom-start',
    });
  }
};
</script>

<ClickOutside on:clickoutside="{hideDropdown}" exclude="{[triggerEl]}">
  <div>
    <a
      class="text-gray-600 block"
      href="##"
      bind:this="{btnDropdownRef}"
      on:click="{toggleDropdown}"
    >
      <div class="items-center flex">
        <span
          class="w-12 h-12 text-sm text-white bg-gray-300 inline-flex items-center justify-center rounded-full"
        >
          <img
            alt="..."
            class="w-full rounded-full align-middle border-none shadow-lg"
            src="https://picsum.photos/seed/todo/256/256"
          />
        </span>
      </div>
    </a>
    <div
      bind:this="{popoverDropdownRef}"
      class="bg-white text-base z-50 float-left py-2 list-none text-left rounded shadow-lg min-w-48 {dropdownPopoverShow ? 'block' : 'hidden'}"
    >
      <button
        on:click="{goToSettings}"
        class="text-sm py-2 px-4 font-normal block w-full whitespace-no-wrap bg-transparent text-gray-800"
      >
        <i class="fa fa-cogs"></i>
        {translationsToUse.settings}
      </button>
      {#if currentAuthStatus.isAdmin && adminMode}
        <div class="h-0 my-2 border border-solid border-gray-200"></div>
        <button
          on:click="{superstore.toggleAdminMode}"
          class="text-sm py-2 px-4 font-normal block w-full whitespace-no-wrap bg-transparent {adminMode ? 'underline text-indigo-500' : ''} "
        >
          <i class="fa fa-user-secret"></i>
          {translationsToUse.adminMode}
          {adminMode ? '✅' : '❌'}
        </button>
      {/if}
      <div class="h-0 my-2 border border-solid border-gray-200"></div>
      <button
        on:click="{logout}"
        class="text-sm py-2 px-4 font-normal block w-full whitespace-no-wrap bg-transparent text-red-600"
      >
        <i class="fa fa-sign-out-alt"></i>
        {translationsToUse.logout}
      </button>
    </div>
  </div>
</ClickOutside>
