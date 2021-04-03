<script lang="typescript">
import { link } from "svelte-routing";

// core components
import Footer from '../components/Footers/MainFooter.svelte';

import { UserSiteSettings } from '../types';
import { sessionSettingsStore } from '../stores';
import {frontendRoutes} from "../constants";

export let location: Location;

let navbarOpen: boolean = false;

function setNavbarOpen() {
  navbarOpen = !navbarOpen;
}

// set up translations
let currentSessionSettings = new UserSiteSettings();
let translationsToUse = currentSessionSettings.getTranslations().pages.home;
const unsubscribeFromSettingsUpdates = sessionSettingsStore.subscribe(
  (value: UserSiteSettings) => {
    currentSessionSettings = value;
    translationsToUse = currentSessionSettings.getTranslations().pages.home;
  },
);
</script>


<nav class="top-0 fixed z-50 w-full flex flex-wrap items-center justify-between px-2 py-3 navbar-expand-lg bg-white shadow">
  <div class="container px-4 mx-auto flex flex-wrap items-center justify-between">
    <div class="w-full relative flex justify-between lg:w-auto lg:static lg:block lg:justify-start">
      <a
        use:link
        class="text-gray-800 text-sm font-bold leading-relaxed inline-block mr-4 py-2 whitespace-no-wrap uppercase"
        href="{frontendRoutes.LANDING}"
      >
        {translationsToUse.navBar.serviceName}
      </a>
      <button
        class="cursor-pointer text-xl leading-none px-3 py-1 border border-solid border-transparent rounded bg-transparent block lg:hidden outline-none focus:outline-none"
        type="button"
        on:click="{setNavbarOpen}"
      >
        <i class="fas fa-bars"></i>
      </button>
    </div>
    <div
      class="lg:flex flex-grow items-center {navbarOpen ? 'block' : 'hidden'}"
      id="example-navbar-warning"
    >
      <ul class="flex flex-col lg:flex-row list-none lg:ml-auto">
        <li class="flex items-right">
          <a
            use:link
            href="{frontendRoutes.LOGIN}"
            class="text-sm py-2 px-4 font-normal block w-full whitespace-no-wrap bg-transparent text-gray-800"
          >
            <button
              class="bg-red-500 text-white active:bg-red-600 text-xs font-bold uppercase px-4 py-2 rounded shadow hover:shadow-lg outline-none focus:outline-none lg:mr-1 lg:mb-0 ml-3 mb-3 ease-linear transition-all duration-150"
              type="button"
            >
              <i class="fas fa-sign-in-alt"></i>&nbsp;&nbsp;{translationsToUse.navBar.buttons.login}
            </button>
          </a>
          <a
            use:link
            href="{frontendRoutes.REGISTER}"
            class="text-sm py-2 px-4 font-normal block w-full whitespace-no-wrap bg-transparent text-gray-800"
          >
            <button
              class="bg-red-500 text-white active:bg-red-600 text-xs font-bold uppercase px-4 py-2 rounded shadow hover:shadow-lg outline-none focus:outline-none lg:mr-1 lg:mb-0 ml-3 mb-3 ease-linear transition-all duration-150"
              type="button"
            >
              <i class="fas fa-user-alt"></i>&nbsp;&nbsp;{translationsToUse.navBar.buttons.register}
            </button>
          </a>
        </li>
      </ul>
    </div>
  </div>
</nav>
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
