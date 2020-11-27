<script lang="typescript">
import { navigate } from 'svelte-routing';

import { Logger } from '../../logger';
import { ErrorResponse, UserSiteSettings } from '../../types';
import { translations } from '../../i18n';
import { sessionSettingsStore } from '../../stores';
import { V1APIClient } from '../../apiClient';
import { frontendRoutes } from '../../constants';
import { AxiosError } from 'axios';

export let location: Location;

let logger = new Logger().withDebugValue(
  'source',
  'src/views/admin/Settings.svelte',
);

// set up translations
let currentSessionSettings = new UserSiteSettings();
let translationsToUse = currentSessionSettings.getTranslations().pages
  .siteSettings;

const unsubscribeFromSettingsUpdates = sessionSettingsStore.subscribe(
  (value: UserSiteSettings) => {
    currentSessionSettings = value;
    translationsToUse = currentSessionSettings.getTranslations().pages
      .siteSettings;
  },
);
// onDestroy(unsubscribeFromSettingsUpdates);

let cookieSecretReplacementError = '';
function confirmCookieSecretReplacement() {
  cookieSecretReplacementError = '';
  if (confirm(translationsToUse.confirmations.cycleCookieSecret)) {
    V1APIClient.cycleCookieSecret()
      .then(() => {
        navigate(frontendRoutes.LOGIN, { state: {}, replace: true });
      })
      .catch((reason: AxiosError<ErrorResponse>) => {
        cookieSecretReplacementError = reason.message;
      });
  }
}
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
        <form>
          <h6 class="text-gray-500 text-sm mt-3 mb-6 font-bold uppercase">
            {translationsToUse.sectionLabels.actions}
          </h6>
          <div class="flex flex-wrap">
            <div class="w-full lg:w-6/12 px-4">
              <button
                class="bg-red-500 text-white active:bg-red-600 font-bold uppercase text-xs px-4 py-2 rounded shadow hover:shadow-md outline-none focus:outline-none mr-1 ease-linear transition-all duration-150"
                on:click="{confirmCookieSecretReplacement}"
                type="button"
              >
                {translationsToUse.buttons.cycleCookieSecret}
              </button>
              <span class="text-red-600">{cookieSecretReplacementError}</span>
            </div>
          </div>
        </form>
      </div>
    </div>
  </div>
</div>
