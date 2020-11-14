import { fetchLanguage, supportedLanguage } from '@/i18n';

import type { LanguageTag } from 'typed-intl';

const defaultLanguage = 'en-US';

export class UserSiteSettings {
  language: LanguageTag;
  darkMode: boolean;

  constructor(language?: supportedLanguage, darkMode: boolean = false) {
    if (!language) {
      switch (window.navigator.language) {
        case 'es-MX':
        case 'es-419':
          language = 'es-MX';
          break;
        default:
          language = defaultLanguage;
          break;
      }
    }

    this.language = fetchLanguage(language);
    this.darkMode = darkMode;
  }
}
