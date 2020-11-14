<script lang="typescript">
  import axios, { AxiosError, AxiosResponse } from 'axios';
  import { link, navigate } from 'svelte-routing';
  import { onDestroy } from 'svelte';

  import { Logger } from '../../logger';
  import { V1APIClient } from '../../requests';
  import {
    RegistrationRequest,
    UserRegistrationResponse,
    TOTPTokenValidationRequest,
    ErrorResponse,
    UserSiteSettings,
  } from '../../types';
  import { translations } from '../../i18n';
  import { sessionSettingsStore } from '../../stores';

  // set up translations
  let currentSessionSettings = new UserSiteSettings();
  let translationsToUse = translations.messagesFor(
    currentSessionSettings.language,
  ).pages.registration;
  const unsubscribeFromSettingsUpdates = sessionSettingsStore.subscribe(
    (value: UserSiteSettings) => {
      currentSessionSettings = value;
      translationsToUse = translations.messagesFor(
        currentSessionSettings.language,
      ).pages.registration;
    },
  );
  // onDestroy(unsubscribeFromSettingsUpdates);

  export let location: Location;

  const registrationRequest = new RegistrationRequest();
  const totpValidationRequest = new TOTPTokenValidationRequest();

  let registrationMayProceed = false;
  let registrationError = '';
  let postRegistrationQRCode = '';
  let totpTokenValidationMayProceed = false;

  let logger = new Logger().withDebugValue(
    'source',
    'src/views/auth/Register.svelte',
  );

  function evaluateCreationInputs(): string {
    const usernameIsLongEnough = registrationRequest.username.length >= 8;
    const passwordsMatch =
      registrationRequest.password === registrationRequest.repeatedPassword;
    const passwordIsLongEnough = registrationRequest.password.length >= 8;

    const reasons: string[] = [];
    if (!usernameIsLongEnough) {
      reasons.push(
        `username '${registrationRequest.username}' has too few characters`,
      );
    } else if (!passwordsMatch) {
      reasons.push("passwords don't match");
    } else if (passwordsMatch && !passwordIsLongEnough) {
      reasons.push(
        `password is not long enough (has ${registrationRequest.password.length})`,
      );
    }

    registrationMayProceed =
      usernameIsLongEnough && passwordIsLongEnough && passwordsMatch;

    let last = reasons.pop();
    if (reasons.length === 1) {
      return last || '';
    } else if (reasons.length > 1) {
      return reasons.join(', ') + ' and ' + last;
    }

    return '';
  }

  async function register() {
    logger.debug('register called');

    if (!registrationMayProceed) {
      // this should never occur
      registrationError = evaluateCreationInputs();
      throw new Error('registration input is not valid!');
    }

    return V1APIClient.registrationRequest(registrationRequest)
      .then((response: AxiosResponse<UserRegistrationResponse>) => {
        const data = response.data;

        postRegistrationQRCode = data.qrCode;
        totpValidationRequest.userID = data.id;

        return data;
      })
      .catch((reason: AxiosError<ErrorResponse>) => {
        if (reason.response) {
          const data = reason.response.data;
          logger.error(data.message);
          registrationError = data.message;
        }
      });
  }

  function evaluateValidationInputs(): void {
    totpTokenValidationMayProceed =
      totpValidationRequest.totpToken.length === 6;
  }

  async function validateTOTPToken() {
    logger.debug('validateTOTPToken called');

    const path = '/users/totp_secret/verify';

    if (!totpTokenValidationMayProceed) {
      // this should never occur
      throw new Error('TOTP token validation input is not valid!');
    }

    return V1APIClient.validateTOTPSecretWithToken(totpValidationRequest)
      .then((response: AxiosResponse) => {
        logger.debug(
          `navigating to /auth/login because totp validation request succeeded`,
        );
        navigate('/auth/login', { state: {}, replace: true });
      })
      .catch((reason: AxiosError<ErrorResponse>) => {
        if (reason.response) {
          const data = reason.response.data;
          logger.error(data.message);
        }
      });
  }
</script>

<div class="container mx-auto px-4 h-full">
  <div class="flex content-center items-center justify-center h-full">
    <div class="w-full lg:w-6/12 px-4">
      <div
        class="relative flex flex-col min-w-0 break-words w-full mb-6 shadow-lg rounded-lg bg-gray-300 border-0">
        <div class="rounded-t mb-0 px-6 py-6" />
        <!-- spacer div -->
        {#if postRegistrationQRCode === ''}
          <div class="flex-auto px-4 lg:px-10 py-10 pt-0">
            <form>
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
                  placeholder={translationsToUse.inputPlaceholders.username}
                  on:keyup={evaluateCreationInputs}
                  on:blur={evaluateCreationInputs}
                  bind:value={registrationRequest.username}
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
                  placeholder={translationsToUse.inputPlaceholders.password}
                  on:keyup={evaluateCreationInputs}
                  on:blur={evaluateCreationInputs}
                  bind:value={registrationRequest.password}
                />
              </div>

              <div class="relative w-full mb-3">
                <label
                  class="block uppercase text-gray-700 text-xs font-bold mb-2"
                  for="passwordRepeatInput"
                >
                  {translationsToUse.inputLabels.passwordRepeat}
                </label>
                <input
                  id="passwordRepeatInput"
                  type="password"
                  class="px-3 py-3 placeholder-gray-400 text-gray-700 bg-white rounded text-sm shadow focus:outline-none focus:shadow-outline w-full ease-linear transition-all duration-150"
                  placeholder={translationsToUse.inputPlaceholders.passwordRepeat}
                  on:keyup={evaluateCreationInputs}
                  on:blur={evaluateCreationInputs}
                  bind:value={registrationRequest.repeatedPassword}
                />
              </div>

              <!--
              <div>
                <label class="inline-flex items-center cursor-pointer">
                  <input
                    id="customCheckLogin"
                    type="checkbox"
                    class="form-checkbox text-gray-800 ml-1 w-5 h-5 ease-linear transition-all duration-150"
                  />
                  <span class="ml-2 text-sm font-semibold text-gray-700">
                    I agree with the
                    <a href="#pablo" on:click={(e) => e.preventDefault()} class="text-red-500">
                      Privacy Policy
                    </a>
                  </span>
                </label>
              </div>
              -->

              <p class="text-red-600">{registrationError}</p>

              <div class="text-center mt-6">
                <button
                  id="registrationButton"
                  class="bg-gray-900 text-white active:bg-gray-700 text-sm font-bold uppercase px-6 py-3 rounded shadow hover:shadow-lg outline-none focus:outline-none mr-1 mb-1 w-full ease-linear transition-all duration-150"
                  type="button"
                  on:click={register}
                >
                  {translationsToUse.buttons.register}
                </button>
              </div>
            </form>
          </div>
        {:else}
          <div class="text-center">
            <img
              id="twoFactorSecretQRCode"
              class="w-1/2 object-center inline p-4"
              src={postRegistrationQRCode}
              alt="two factor authentication secret encoded as a QR code"
            >
            <p class="m-4">
              {translationsToUse.notices.saveQRSecretNotice}
            </p>
            <p class="p-4">
              {translationsToUse.instructions.enterGeneratedTwoFactorCode}
              <input
                id="totpTokenInput"
                bind:value={totpValidationRequest.totpToken}
                type="text"
                placeholder={translationsToUse.inputPlaceholders.twoFactorCode}
                on:keyup={evaluateValidationInputs}
                on:blur={evaluateValidationInputs}
              >
            </p>
            <p class="p-4">
              <button
                id="totpTokenSubmitButton"
                class="bg-gray-900 text-white active:bg-gray-700 text-sm font-bold uppercase px-6 py-3 rounded shadow hover:shadow-lg outline-none focus:outline-none mr-1 mb-1 w-full ease-linear transition-all duration-150"
                type="button"
                on:click={validateTOTPToken}
              >
                {translationsToUse.buttons.submitVerification}
              </button>
            </p>
          </div>
        {/if}
      </div>
      <div class="flex flex-wrap mt-6 relative">
        <div class="w-1/2" />
        <div class="w-1/2 text-right">
          <a use:link href="/auth/login" class="text-gray-300">
            <small>{translationsToUse.linkTexts.loginInstead}</small>
          </a>
        </div>
      </div>
    </div>
  </div>
</div>
