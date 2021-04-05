<script lang="typescript">
import { AxiosError, AxiosResponse } from 'axios';
import { link, navigate } from 'svelte-routing';

import { UserStatus, LoginRequest } from '../../types';
import { V1APIClient } from '../../apiClient';
import { Logger } from '../../logger';
import { UserSiteSettings } from '../../types';
import { frontendRoutes } from '../../constants';
import { Superstore } from '../../stores';

export let location: Location;

// set up translations
let currentSessionSettings = new UserSiteSettings();
let translationsToUse = currentSessionSettings.getTranslations().pages.login;
const superstore = new Superstore({
  sessionSettingsStoreUpdateFunc: (value: UserSiteSettings) => {
    currentSessionSettings = value;
    translationsToUse = currentSessionSettings.getTranslations().pages.login;
  },
});

let logger = new Logger().withDebugValue(
  'source',
  'src/views/auth/login.svelte',
);

const loginRequest = new LoginRequest();
let canLogin: boolean = false;
let canPressLoginButton: boolean = canLogin;
let loginError: string = '';

function evaluateInputs(): void {
  canLogin =
    loginRequest.username !== '' &&
    loginRequest.password !== '' &&
    loginRequest.totpToken.length > 0 &&
    loginRequest.totpToken.length <= 6;
  canPressLoginButton = canLogin
}

async function login() {
  logger.debug('login called!');
  evaluateInputs();

  if (!canLogin) {
    throw new Error('invalid input!');
  }

  canPressLoginButton = false;

  if (superstore.frontendOnlyMode) {
    navigate(frontendRoutes.ADMIN_DASHBOARD, {
      state: {},
      replace: true,
    });
  } else {
    return V1APIClient.login(loginRequest)
      .then((res: AxiosResponse<UserStatus>) => {
        const userStatus: UserStatus = res.data;
        superstore.setUserStatus(userStatus);

        navigate(frontendRoutes.ADMIN_DASHBOARD, { state: {}, replace: true });
      })
      .catch((reason: AxiosError) => {
        if (reason?.response?.status === 401) {
            loginError = 'invalid credentials: please try again';
            canPressLoginButton = true;
          } else {
            loginError = reason.response.toString();
            logger.error(JSON.stringify(reason.response));
            canPressLoginButton = true;
          }
      });
  }
}
</script>

<div class="container mx-auto px-4 h-full">
  <div class="flex content-center items-center justify-center h-full">
    <div class="w-full lg:w-4/12 px-4">
      <div class="relative flex flex-col min-w-0 break-words w-full mb-6 rounded-lg border-0">
        <div class="rounded-t mb-0 px-6 py-6"></div>
        <div class="flex-auto px-4 lg:px-10 py-10 pt-0">
          <form on:submit|preventDefault="{login}">
            <div class="relative w-full mb-3">
              <label
                class="block uppercase text-white text-xs font-bold mb-2"
                for="usernameInput"
              >
                {translationsToUse.inputLabels.username}
              </label>
              <input
                id="usernameInput"
                tabindex="0"
                type="text"
                class="px-3 py-3 placeholder-gray-400 text-black bg-white rounded text-sm shadow focus:outline-none focus:shadow-outline w-full ease-linear transition-all duration-150"
                placeholder="{translationsToUse.inputPlaceholders.username}"
                on:keyup="{evaluateInputs}"
                on:blur="{evaluateInputs}"
                bind:value="{loginRequest.username}"
              />
            </div>

            <div class="relative w-full mb-3">
              <label
                class="block uppercase text-white text-xs font-bold mb-2"
                for="passwordInput"
              >
                {translationsToUse.inputLabels.password}
              </label>
              <input
                id="passwordInput"
                type="password"
                class="px-3 py-3 placeholder-gray-400 text-black bg-white rounded text-sm shadow focus:outline-none focus:shadow-outline w-full ease-linear transition-all duration-150"
                placeholder="{translationsToUse.inputPlaceholders.password}"
                on:keyup="{evaluateInputs}"
                on:blur="{evaluateInputs}"
                bind:value="{loginRequest.password}"
              />
            </div>

            <div class="relative w-full mb-3">
              <label
                class="block uppercase text-white text-xs font-bold mb-2"
                for="totpTokenInput"
              >
                {translationsToUse.inputLabels.twoFactorCode}
              </label>
              <input
                id="totpTokenInput"
                type="text"
                class="px-3 py-3 placeholder-gray-400 text-black bg-white rounded text-sm shadow focus:outline-none focus:shadow-outline w-full ease-linear transition-all duration-150"
                placeholder="{translationsToUse.inputPlaceholders.twoFactorCode}"
                on:keyup="{evaluateInputs}"
                on:blur="{evaluateInputs}"
                bind:value="{loginRequest.totpToken}"
              />
            </div>

            {#if loginError !== ''}
              <p class="text-red-600">{loginError}</p>
            {/if}

            <div class="text-center mt-6">
              <button
                on:mouseup="{login}"
                  disabled={!canPressLoginButton}
                type="submit"
                id="loginButton"
                class="{canLogin ? 'bg-gray-900 active:bg-gray-700 text-white' : 'bg-gray-300 text-black'} active:bg-gray-700 text-sm font-bold uppercase px-6 py-3 border border-white rounded shadow hover:shadow-lg outline-none focus:outline-none mr-1 mb-1 w-full ease-linear transition-all duration-150"
              >
                {translationsToUse.buttons.login}
              </button>
            </div>
          </form>
        </div>
      </div>
      <div class="flex flex-wrap mt-6 relative">
        <div class="w-1/2">
          <a
            href="##"
            on:click="{(e) => e.preventDefault()}"
            class="text-gray-300"
          >
            <small>{translationsToUse.linkTexts.forgotPassword}</small>
          </a>
        </div>
        <div class="w-1/2 text-right">
          <a use:link href="{frontendRoutes.REGISTER}" class="text-gray-300">
            <small>{translationsToUse.linkTexts.createAccount}</small>
          </a>
        </div>
      </div>
    </div>
  </div>
</div>
