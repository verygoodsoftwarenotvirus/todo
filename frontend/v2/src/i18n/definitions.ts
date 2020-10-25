import {LanguageTag, languageTag, translate} from "typed-intl";

import {englishTranslations} from "@/i18n/english";
import type {
  adminNavbarTranslations,
  authNavbarTranslations,
  homepageNavbarTranslations,
  userDropdownTranslations,
  itemModelTranslations,
  homePageTranslations,
  loginPageTranslations,
  registrationPageTranslations,
  userSettingsPageTranslations,
  userModelTranslations,
  apiTableTranslations,
  primarySidebarTranslations,
  mainFooterTranslations,
  adminFooterTranslations,
  smallFooterTranslations,
} from "@/i18n";

const english = "en-US";
const mexicanSpanish = "es-MX";
const defaultLanguage = english;

export type supportedLanguage = "en-US" | "es-MX";

export function fetchLanguage(name: supportedLanguage): LanguageTag {
  switch (name.toLowerCase().trim()) {
  case mexicanSpanish.toLowerCase().trim():
    return languageTag(mexicanSpanish)
  default:
    return languageTag(defaultLanguage)
  }
}

export type SiteTranslationMap = {
  components: {
    apiTable: apiTableTranslations,
    dropdowns: {
      userDropdown: userDropdownTranslations,
    },
    navbars: {
      adminNavbar: adminNavbarTranslations,
      authNavbar: authNavbarTranslations,
      homepageNavbar:homepageNavbarTranslations,
    },
    sidebars: {
      primary: primarySidebarTranslations,
    },
    footers: {
      mainFooter: mainFooterTranslations,
      adminFooter: adminFooterTranslations,
      smallFooter: smallFooterTranslations,
    },
  },
  pages: {
    home: homePageTranslations,
    login: loginPageTranslations,
    registration: registrationPageTranslations,
    userSettings: userSettingsPageTranslations,
  },
  models: {
    item: itemModelTranslations,
    user: userModelTranslations,
  },
}

export const translations = translate(
  englishTranslations,
)
// .supporting(mexicanSpanish,
//     // TODO: actually translate, lol
//     {},
// )
