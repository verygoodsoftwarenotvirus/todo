<script lang="typescript">
  // make dynamic date to be added to footer
  import { UserSiteSettings } from '../../types';
  import { translations } from '../../i18n';
  import { sessionSettingsStore } from '../../stores';
  import { onDestroy } from 'svelte';

  // set up translations
  let currentSessionSettings = new UserSiteSettings();
  let translationsToUse = translations.messagesFor(
    currentSessionSettings.language,
  ).components.footers.adminFooter;
  const unsubscribeFromSettingsUpdates = sessionSettingsStore.subscribe(
    (value: UserSiteSettings) => {
      currentSessionSettings = value;
      translationsToUse = translations.messagesFor(
        currentSessionSettings.language,
      ).components.footers.adminFooter;
    },
  );
  // onDestroy(unsubscribeFromSettingsUpdates);
</script>

<footer class="block py-4">
  <div class="container mx-auto px-4">
    <hr class="mb-4 border-b-1 border-gray-300" />
    <div class="flex flex-wrap items-center md:justify-between justify-center">
      <div class="w-full md:w-4/12 px-4">
        <div
          class="text-sm text-gray-600 font-semibold py-1 text-center md:text-left">
          {translationsToUse.copyright}
          {new Date().getFullYear()}
        </div>
      </div>
      <div class="w-full md:w-8/12 px-4">
        <ul class="flex flex-wrap list-none md:justify-end justify-center">
          <li>
            <a
              href="##"
              class="text-gray-700 hover:text-gray-900 text-sm font-semibold block py-1 px-3">
              {translationsToUse.aboutUs}
            </a>
          </li>
          <li>
            <a
              href="##"
              class="text-gray-700 hover:text-gray-900 text-sm font-semibold block py-1 px-3">
              {translationsToUse.blog}
            </a>
          </li>
        </ul>
      </div>
    </div>
  </div>
</footer>
