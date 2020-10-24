import {fetchLanguage, supportedLanguage} from "@/i18n";

import type {LanguageTag} from "typed-intl";

const defaultLanguage = "en"

export class SessionSettings {
    language: LanguageTag;

    constructor(language: supportedLanguage = defaultLanguage) {
        this.language = fetchLanguage(language);
    }
}
