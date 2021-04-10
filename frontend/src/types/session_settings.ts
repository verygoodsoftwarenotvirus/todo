import type { LanguageTag } from 'typed-intl';
import {
  fetchLanguage,
  SiteTranslationMap,
  supportedLanguage,
  translations,
} from '../i18n';

const defaultLanguage = 'en-US';

export class UserSiteSettings {
  language: LanguageTag;
  darkMode: boolean;

  constructor(language?: supportedLanguage, darkMode: boolean = false) {
    this.darkMode = darkMode;

    if (!language) {
      switch (window.navigator.language) {
        // case 'es-MX':
        // case 'es-419':
        //   language = 'es-MX';
        //   break;
        default:
          language = defaultLanguage;
          break;
      }
    }

    this.language = fetchLanguage(language);
  }

  getTranslations(): Readonly<SiteTranslationMap> {
    return translations.messagesFor(this.language);
  }
}
