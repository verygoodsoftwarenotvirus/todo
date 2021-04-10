import type {
  accountModelTranslations,
  accountSettingsPageTranslations,
  adminFooterTranslations,
  adminNavbarTranslations,
  apiClientModelTranslations,
  apiTableTranslations,
  auditLogEntryTableTranslations,
  auditLogEntryTranslations,
  authNavbarTranslations,
  homePageTranslations,
  itemModelTranslations,
  loginPageTranslations,
  mainFooterTranslations,
  primarySidebarTranslations,
  registrationPageTranslations,
  siteSettingsPageTranslations,
  smallFooterTranslations,
  userAdminPageTranslations,
  userDropdownTranslations,
  userModelTranslations,
  userSettingsPageTranslations,
  webhookCreationPageTranslations,
  webhookModelTranslations,
} from '../i18n';
import { englishTranslations } from '../i18n/english';
import { LanguageTag, languageTag, translate } from 'typed-intl';

const english = 'en-US';
const mexicanSpanish = 'es-MX';
export type supportedLanguage = 'en-US'; // | 'es-MX';

const defaultLanguage = english;

export function fetchLanguage(name: supportedLanguage): LanguageTag {
  switch (name.toLowerCase().trim()) {
    case mexicanSpanish.toLowerCase().trim():
      return languageTag(mexicanSpanish);
    default:
      return languageTag(defaultLanguage);
  }
}

export type SiteTranslationMap = {
  components: {
    apiTable: apiTableTranslations;
    dropdowns: {
      userDropdown: userDropdownTranslations;
    };
    auditLogEntryTable: auditLogEntryTableTranslations;
    navbars: {
      adminNavbar: adminNavbarTranslations;
      authNavbar: authNavbarTranslations;
    };
    sidebars: {
      primary: primarySidebarTranslations;
    };
    footers: {
      mainFooter: mainFooterTranslations;
      adminFooter: adminFooterTranslations;
      smallFooter: smallFooterTranslations;
    };
  };
  pages: {
    home: homePageTranslations;
    login: loginPageTranslations;
    registration: registrationPageTranslations;
    accountSettings: accountSettingsPageTranslations;
    userSettings: userSettingsPageTranslations;
    siteSettings: siteSettingsPageTranslations;
    webhookCreationPage: webhookCreationPageTranslations;
    userAdminPage: userAdminPageTranslations;
  };
  models: {
    item: itemModelTranslations;
    account: accountModelTranslations;
    apiClient: apiClientModelTranslations;
    user: userModelTranslations;
    auditLogEntry: auditLogEntryTranslations;
    webhook: webhookModelTranslations;
  };
};

export const translations = translate(englishTranslations);
// .supporting(mexicanSpanish,
//     // TODO: actually translate, lol
//     {},
// )
