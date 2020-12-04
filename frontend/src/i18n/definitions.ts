import type {
  adminFooterTranslations,
  adminNavbarTranslations,
  apiTableTranslations,
  auditLogEntryTableTranslations,
  auditLogEntryTranslations,
  authNavbarTranslations,
  homepageNavbarTranslations,
  homePageTranslations,
  itemModelTranslations,
  loginPageTranslations,
  mainFooterTranslations,
  oauth2ClientModelTranslations,
  primarySidebarTranslations,
  registrationPageTranslations,
  siteSettingsPageTranslations,
  smallFooterTranslations,
  userAdminPageTranslations,
  userDropdownTranslations,
  userModelTranslations,
  userSettingsPageTranslations,
  webhookModelTranslations,
} from '@/i18n';
import { englishTranslations } from '@/i18n/english';
import { LanguageTag, languageTag, translate } from 'typed-intl';

const english = 'en-US';
const mexicanSpanish = 'es-MX';
export type supportedLanguage = 'en-US' | 'es-MX';

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
      homepageNavbar: homepageNavbarTranslations;
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
    userAdminPageTranslations: userAdminPageTranslations;
    registration: registrationPageTranslations;
    userSettings: userSettingsPageTranslations;
    siteSettings: siteSettingsPageTranslations;
  };
  models: {
    item: itemModelTranslations;
    user: userModelTranslations;
    auditLogEntry: auditLogEntryTranslations;
    oauth2Client: oauth2ClientModelTranslations;
    webhook: webhookModelTranslations;
  };
};

export const translations = translate(englishTranslations);
// .supporting(mexicanSpanish,
//     // TODO: actually translate, lol
//     {},
// )
