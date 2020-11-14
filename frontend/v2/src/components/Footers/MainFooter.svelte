<script lang="typescript">
  import { onDestroy } from 'svelte';

  import { UserSiteSettings } from '../../types';
  import { translations } from '../../i18n';
  import { sessionSettingsStore } from '../../stores';

  let currentSessionSettings = new UserSiteSettings();
  let translationsToUse = translations.messagesFor(
    currentSessionSettings.language,
  ).components.footers.mainFooter;
  const unsubscribeFromSettingsUpdates = sessionSettingsStore.subscribe(
    (value: UserSiteSettings) => {
      currentSessionSettings = value;
      translationsToUse = translations.messagesFor(
        currentSessionSettings.language,
      ).components.footers.mainFooter;
    },
  );
  // onDestroy(unsubscribeFromSettingsUpdates);
</script>

<footer class="relative bg-gray-300 pt-8 pb-6">
  <div
    class="bottom-auto top-0 left-0 right-0 w-full absolute pointer-events-none overflow-hidden -mt-20 h-20"
    style="transform: translateZ(0);">
    <svg
      class="absolute bottom-0 overflow-hidden"
      xmlns="http://www.w3.org/2000/svg"
      preserveAspectRatio="none"
      version="1.1"
      viewBox="0 0 2560 100"
      x="0"
      y="0">
      <polygon
        class="text-gray-300 fill-current"
        points="2560 0 2560 100 0 100" />
    </svg>
  </div>
  <div class="container mx-auto px-4">
    <div class="flex flex-wrap text-center lg:text-left">
      <div class="w-full lg:w-6/12 px-4">
        <h4 class="text-3xl font-semibold">{translationsToUse.keepInTouch}</h4>
        <h5 class="text-lg mt-0 mb-2 text-gray-700">
          {translationsToUse.weLikeYou}
        </h5>
      </div>
      <div class="w-full lg:w-6/12 px-4">
        <div class="flex flex-wrap items-top mb-6">
          <div class="w-full lg:w-4/12 px-4 ml-auto">
            <span
              class="block uppercase text-gray-600 text-sm font-semibold mb-2">
              {translationsToUse.usefulLinks}
            </span>
            <ul class="list-unstyled">
              <li>
                <a
                  class="text-gray-700 hover:text-gray-900 font-semibold block pb-2 text-sm"
                  href="##">
                  {translationsToUse.aboutUs}
                </a>
              </li>
              <li>
                <a
                  class="text-gray-700 hover:text-gray-900 font-semibold block pb-2 text-sm"
                  href="##">
                  {translationsToUse.blog}
                </a>
              </li>
            </ul>
          </div>
          <div class="w-full lg:w-4/12 px-4">
            <span
              class="block uppercase text-gray-600 text-sm font-semibold mb-2">
              {translationsToUse.otherResources}
            </span>
            <ul class="list-unstyled">
              <li>
                <a
                  class="text-gray-700 hover:text-gray-900 font-semibold block pb-2 text-sm"
                  href="##">
                  {translationsToUse.termsAndConditions}
                </a>
              </li>
              <li>
                <a
                  class="text-gray-700 hover:text-gray-900 font-semibold block pb-2 text-sm"
                  href="##">
                  {translationsToUse.privacyPolicy}
                </a>
              </li>
              <li>
                <a
                  class="text-gray-700 hover:text-gray-900 font-semibold block pb-2 text-sm"
                  href="##">
                  {translationsToUse.contactUs}
                </a>
              </li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  </div>
</footer>
