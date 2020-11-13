<script lang="typescript">
  import {onDestroy} from "svelte";

  // core components
  import IndexNavbar from "../components/Navbars/IndexNavbar.svelte";
  import Footer from "../components/Footers/MainFooter.svelte";

  import {UserSiteSettings} from "../types";
  import {translations} from "../i18n";
  import {sessionSettingsStore} from "../stores";

  export let location: Location;

  // set up translations
  let currentSessionSettings = new UserSiteSettings();
  let translationsToUse = translations.messagesFor(currentSessionSettings.language).pages.home;
  const unsubscribeFromSettingsUpdates = sessionSettingsStore.subscribe((value: UserSiteSettings) => {
    currentSessionSettings = value;
    translationsToUse = translations.messagesFor(currentSessionSettings.language).pages.home;
  });
  onDestroy(unsubscribeFromSettingsUpdates);
</script>

<IndexNavbar />
<section class="header relative pt-16 items-center flex h-screen max-h-860-px">
  <div class="container mx-auto items-center flex flex-wrap">
    <div class="w-full md:w-8/12 lg:w-6/12 xl:w-6/12 px-4">
      <div class="pt-32 sm:pt-0">
        <h2 class="font-semibold text-4xl text-gray-700">
          {translationsToUse.mainGreeting}
        </h2>
        <p class="mt-4 text-lg leading-relaxed text-gray-600">
          {translationsToUse.subGreeting}
        </p>
      </div>
    </div>
  </div>
</section>
<Footer />
