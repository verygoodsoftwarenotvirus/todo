<script lang="typescript">
import { onDestroy, onMount } from 'svelte';
import axios, { AxiosError, AxiosResponse } from 'axios';
import { navigate } from 'svelte-routing';

import { userStatusStore } from '../../stores';
import {
  User,
  UserStatus,
  UserPasswordUpdateRequest,
  UserTwoFactorSecretUpdateRequest,
  ErrorResponse,
  UserSiteSettings,
} from '../../types';
import { Logger } from '../../logger';
import { V1APIClient } from '../../apiClient';
import { translations } from '../../i18n';
import { sessionSettingsStore } from '../../stores';
import { frontendRoutes } from '../../constants';
import { Superstore } from '../../stores';

export let location: Location;

let logger = new Logger().withDebugValue(
  'source',
  'src/views/user/Settings.svelte',
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

let ogUser: User = new User();
let user: User = new User();
let passwordUpdate = new UserPasswordUpdateRequest();
let twoFactorSecretUpdate = new UserTwoFactorSecretUpdateRequest();

let userInfoCanBeSaved: boolean = false;
let userFetchError: string = '';
let updatePasswordError: string = '';
let twoFactorSecretUpdateError: string = '';

function submitChangePasswordRequest() {
  logger.debug(`submitChangePasswordRequest invoked`);

  V1APIClient.passwordChangeRequest(passwordUpdate)
    .then((res: AxiosResponse) => {
      logger
        .withValue('responseData', res.data)
        .info('passwordChangeRequest returned');
      navigate(frontendRoutes.LOGIN, { state: {}, replace: false });
    })
    .catch((err: AxiosError<ErrorResponse>) => {
      logger.error(err.message);
      updatePasswordError = err.message;
    });
}

onMount(() => {
  V1APIClient.selfRequest()
    .then((resp: AxiosResponse<User>) => {
      user = resp.data;
      ogUser = { ...user };
    })
    .catch((err: AxiosError<ErrorResponse>) => {
      userFetchError = err.message;
    });
});
</script>

<div class="flex flex-wrap">
  <div class="w-full px-4">
    <div
      class="relative flex flex-col min-w-0 break-words w-full mb-6 shadow-lg rounded-lg bg-gray-200 border-0"
    >
      <div class="rounded-t bg-white mb-0 px-6 py-6">
        <div class="text-center flex justify-between">
          <h6 class="text-gray-800 text-xl font-bold">
            {translationsToUse.title}
          </h6>
        </div>
      </div>

      <div class="flex-auto px-4 lg:px-10 py-10 pt-0">
        <div class="text-center flex justify-between">
          <h6 class="text-gray-500 text-sm mt-3 mb-6 font-bold uppercase">
            {translationsToUse.sectionLabels.userInfo}
          </h6>
          <button
            class="{passwordUpdate.goodToGo() ? 'bg-blue-500' : 'bg-gray-300'} text-white active:bg-blue-600 font-bold uppercase text-xs rounded p-3 m-2"
            type="button"
          >
            {translationsToUse.buttons.updateUserInfo}
          </button>
        </div>
        <div class="flex flex-wrap">
          <div class="w-full lg:w-6/12 px-4">
            <div class="relative w-full mb-3">
              <label
                class="block uppercase text-gray-700 text-xs font-bold mb-2"
                for="grid-username"
              >
                {translationsToUse.inputLabels.username}
              </label>
              <input
                id="grid-username"
                type="text"
                disabled
                class="px-3 py-3 placeholder-gray-400 text-gray-700 bg-gray-300 rounded text-sm shadow focus:outline-none focus:shadow-outline w-full ease-linear transition-all duration-150"
                bind:value="{user.username}"
              />
            </div>
          </div>
          <div class="w-full lg:w-6/12 px-4">
            <div class="relative w-full mb-3">
              <label
                class="block uppercase text-gray-700 text-xs font-bold mb-2"
                for="grid-email"
              >
                {translationsToUse.inputLabels.emailAddress}
              </label>
              <input
                id="grid-email"
                type="email"
                class="px-3 py-3 placeholder-gray-400 text-gray-700 bg-gray-300 rounded text-sm shadow focus:outline-none focus:shadow-outline w-full ease-linear transition-all duration-150"
                disabled
                value="we don't want your stinkin' email"
              />
            </div>
          </div>
        </div>

        <hr class="mt-6 border-b-1 border-gray-400" />

        <div class="text-center flex justify-between">
          <h6 class="text-gray-500 text-sm mt-3 mb-6 font-bold uppercase">
            {translationsToUse.sectionLabels.password}
          </h6>
          <button
            on:click="{submitChangePasswordRequest}"
            class="{passwordUpdate.goodToGo() ? 'bg-blue-500' : 'bg-gray-300'} text-white active:bg-blue-600 font-bold uppercase text-xs rounded p-3 m-2"
            type="button"
          >
            {translationsToUse.buttons.changePassword}
          </button>
        </div>
        <div class="flex flex-wrap">
          <div class="w-full lg:w-4/12 px-4">
            <div class="relative w-full mb-3">
              <label
                class="block uppercase text-gray-700 text-xs font-bold mb-2"
                for="grid-password-update-current-password"
              >
                {translationsToUse.inputLabels.currentPassword}
              </label>
              <input
                id="grid-password-update-current-password"
                type="password"
                placeholder="{translationsToUse.inputPlaceholders.currentPassword}"
                class="px-3 py-3 placeholder-gray-400 text-gray-700 bg-white rounded text-sm shadow focus:outline-none focus:shadow-outline w-full ease-linear transition-all duration-150"
                bind:value="{passwordUpdate.currentPassword}"
              />
            </div>
          </div>
          <div class="w-full lg:w-4/12 px-4">
            <div class="relative w-full mb-3">
              <label
                class="block uppercase text-gray-700 text-xs font-bold mb-2"
                for="grid-password-update-new-password"
              >
                {translationsToUse.inputLabels.newPassword}
              </label>
              <input
                id="grid-password-update-new-password"
                type="password"
                placeholder="{translationsToUse.inputPlaceholders.newPassword}"
                class="px-3 py-3 placeholder-gray-400 text-gray-700 bg-white rounded text-sm shadow focus:outline-none focus:shadow-outline w-full ease-linear transition-all duration-150"
                bind:value="{passwordUpdate.newPassword}"
              />
            </div>
          </div>
          <div class="w-full lg:w-4/12 px-4">
            <div class="relative w-full mb-3">
              <label
                class="block uppercase text-gray-700 text-xs font-bold mb-2"
                for="grid-password-update-totp-token"
              >
                {translationsToUse.inputLabels.twoFactorToken}
              </label>
              <input
                id="grid-password-update-totp-token"
                type="text"
                placeholder="{translationsToUse.inputPlaceholders.twoFactorToken}"
                class="px-3 py-3 placeholder-gray-400 text-gray-700 bg-white rounded text-sm shadow focus:outline-none focus:shadow-outline w-full ease-linear transition-all duration-150"
                bind:value="{passwordUpdate.totpToken}"
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</div>
