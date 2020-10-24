<script lang="typescript">
  import {onDestroy} from "svelte";

  import {SessionSettings} from "../../models";
  import {translations} from "../../i18n";
  import {sessionSettingsStore} from "../../stores";

  export let absolute: Boolean = false;

  // set up translations
  let currentSessionSettings = new SessionSettings();
  let translationsToUse = translations.messagesFor(currentSessionSettings.language).components.footers.smallFooter;
  const unsubscribeFromSettingsUpdates = sessionSettingsStore.subscribe((value: SessionSettings) => {
    currentSessionSettings = value;
    translationsToUse = translations.messagesFor(currentSessionSettings.language).components.footers.smallFooter;
  });
  onDestroy(unsubscribeFromSettingsUpdates);
</script>

<footer
  class="pb-6 {absolute ? 'absolute w-full bottom-0 bg-gray-900' : 'relative'}"
>
  <div class="container mx-auto px-4">
    <hr class="mb-6 border-b-1 border-gray-700" />
    <div class="flex flex-wrap items-center md:justify-between justify-center">
      <div class="w-full md:w-4/12 px-4">
        <div
          class="text-sm text-gray-600 font-semibold py-1 text-center md:text-left"
        >
          {translationsToUse.copyright} {new Date().getFullYear()}
        </div>
      </div>
      <div class="w-full md:w-8/12 px-4">
        <ul class="flex flex-wrap list-none md:justify-end justify-center">
          <li>
            <a
              href="#"
              class="text-white hover:text-gray-400 text-sm font-semibold block py-1 px-3"
            >
              {translationsToUse.aboutUs}
            </a>
          </li>
          <li>
            <a
              href="#"
              class="text-white hover:text-gray-400 text-sm font-semibold block py-1 px-3"
            >
              {translationsToUse.blog}
            </a>
          </li>
        </ul>
      </div>
    </div>
  </div>
</footer>
