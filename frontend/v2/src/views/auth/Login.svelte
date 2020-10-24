<script lang="typescript">
  import axios, {AxiosError, AxiosResponse} from 'axios';
  import { link, navigate } from "svelte-routing";

  import { UserStatus, LoginRequest } from "../../models";
  import { userStatusStore, sessionSettingsStore } from "../../stores"
  import { V1APIClient } from "../../requests";
  import { Logger } from "../../logger";
  import {translations} from "../../i18n";
  import {SessionSettings} from "../../models";

  export let location: Location;

  // set up translations
  let currentSessionSettings = new SessionSettings();
  let translationsToUse = translations.messagesFor(currentSessionSettings.language).pages.login;
  const unsubscribeFromSettingsUpdates = sessionSettingsStore.subscribe((value: SessionSettings) => {
    currentSessionSettings = value;
    translationsToUse = translations.messagesFor(currentSessionSettings.language).pages.login;
  });

  let logger = new Logger().withDebugValue("source", "src/views/auth/Login.svelte");

  const loginRequest = new LoginRequest()
  let canLogin: boolean = false;
  let loginError: string = '';

  function evaluateInputs(): void {
    canLogin = loginRequest.username !== '' &&
               loginRequest.password !== '' &&
               loginRequest.totpToken.length > 0 &&
               loginRequest.totpToken.length <= 6;
  }

  async function login() {
    logger.debug("login called!");
    evaluateInputs();

    if (!canLogin) {
      throw new Error("invalid input!");
    }

    return V1APIClient.login(loginRequest)
      .then((res: AxiosResponse<UserStatus>) => {
        const userStatus: UserStatus = res.data;
        userStatusStore.setUserStatus(userStatus);

        if (userStatus.isAdmin) {
          logger.debug(`navigating to /admin/dashboard because user is an authenticated admin`);
          navigate("/admin/dashboard", { state: {}, replace: true });
        } else {
          logger.debug(`navigating to homepage because user is a plain user`);
          navigate("/", { state: {}, replace: true });
        }
      })
      .catch((reason: AxiosError) => {
        if (reason.response) {
          if (reason.response.status === 401) {
            loginError = 'invalid credentials: please try again'
          } else {
            loginError = reason.response.toString();
            logger.error(JSON.stringify(reason.response));
          }
        }
      });
  }
</script>

<div class="container mx-auto px-4 h-full">
  <div class="flex content-center items-center justify-center h-full">
    <div class="w-full lg:w-4/12 px-4">
      <div class="relative flex flex-col min-w-0 break-words w-full mb-6 shadow-lg rounded-lg bg-gray-300 border-0">
        <div class="rounded-t mb-0 px-6 py-6"></div>
        <div class="flex-auto px-4 lg:px-10 py-10 pt-0">
          <form on:submit|preventDefault="{login}">
            <div class="relative w-full mb-3">
              <label
                class="block uppercase text-gray-700 text-xs font-bold mb-2"
                for="usernameInput"
              >
                {translationsToUse.inputLabels.username}
              </label>
              <input
                id="usernameInput"
                type="text"
                class="px-3 py-3 placeholder-gray-400 text-gray-700 bg-white rounded text-sm shadow focus:outline-none focus:shadow-outline w-full ease-linear transition-all duration-150"
                placeholder="{translationsToUse.inputPlaceholders.username}"
                on:keyup={evaluateInputs}
                on:blur={evaluateInputs}
                bind:value={loginRequest.username}
              />
            </div>

            <div class="relative w-full mb-3">
              <label
                class="block uppercase text-gray-700 text-xs font-bold mb-2"
                for="passwordInput"
              >
                {translationsToUse.inputLabels.password}
              </label>
              <input
                id="passwordInput"
                type="password"
                class="px-3 py-3 placeholder-gray-400 text-gray-700 bg-white rounded text-sm shadow focus:outline-none focus:shadow-outline w-full ease-linear transition-all duration-150"
                placeholder="{translationsToUse.inputPlaceholders.password}"
                on:keyup={evaluateInputs}
                on:blur={evaluateInputs}
                bind:value={loginRequest.password}
              />
            </div>

            <div class="relative w-full mb-3">
              <label
                class="block uppercase text-gray-700 text-xs font-bold mb-2"
                for="totpTokenInput"
              >
                {translationsToUse.inputLabels.twoFactorCode}
              </label>
              <input
                id="totpTokenInput"
                type="text"
                class="px-3 py-3 placeholder-gray-400 text-gray-700 bg-white rounded text-sm shadow focus:outline-none focus:shadow-outline w-full ease-linear transition-all duration-150"
                placeholder="{translationsToUse.inputPlaceholders.twoFactorCode}"
                on:keyup={evaluateInputs}
                on:blur={evaluateInputs}
                bind:value={loginRequest.totpToken}
              />
            </div>

            {#if loginError !== ''}
            <p class="text-red-600">{loginError}</p>
            {/if}

            <div class="text-center mt-6">
              <button
                on:click={login}
                type="submit"
                id="loginButton"
                class="bg-gray-900 text-white active:bg-gray-700 text-sm font-bold uppercase px-6 py-3 rounded shadow hover:shadow-lg outline-none focus:outline-none mr-1 mb-1 w-full ease-linear transition-all duration-150"
              >
                {translationsToUse.buttons.login}
              </button>
            </div>
          </form>
        </div>
      </div>
      <div class="flex flex-wrap mt-6 relative">
        <div class="w-1/2">
          <a href="##" on:click={(e) => e.preventDefault()} class="text-gray-300">
            <small>{translationsToUse.linkTexts.forgotPassword}</small>
          </a>
        </div>
        <div class="w-1/2 text-right">
          <a use:link href="/auth/register" class="text-gray-300">
            <small>{translationsToUse.linkTexts.createAccount}</small>
          </a>
        </div>
      </div>
    </div>
  </div>
</div>
