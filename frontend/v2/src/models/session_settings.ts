import {fetchLanguage, supportedLanguage} from "@/i18n";

import type {LanguageTag} from "typed-intl";

const defaultLanguage = "en-US"

export class SessionSettings {
    language: LanguageTag;

    constructor(language?: supportedLanguage) {
        if (!language) {
            switch (window.navigator.language) {
            case "es-MX":
            case "es-419":
                language = "es-MX";
                break;
            default:
                language = defaultLanguage;
                break;
            }
        }
        
        this.language = fetchLanguage(language);
    }
}
