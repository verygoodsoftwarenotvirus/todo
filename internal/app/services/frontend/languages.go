package frontend

import (
	"net/http"

	"golang.org/x/text/language"
)

type (
	displayLanguage *string

	languageDetail struct {
		Tag          language.Tag
		Name         string
		Abbreviation string
	}
)

var (
	english        = new(displayLanguage)
	englishDetails = &languageDetail{
		Name:         "English",
		Abbreviation: "en-US",
		Tag:          language.AmericanEnglish,
	}

	spanish        = new(displayLanguage)
	spanishDetails = &languageDetail{
		Name:         "Spanish",
		Abbreviation: "es-419",
		Tag:          language.LatinAmericanSpanish,
	}

	languageDetails = map[*displayLanguage]*languageDetail{
		english: englishDetails,
		spanish: spanishDetails,
	}
	supportedLanguages = []*displayLanguage{
		english,
		spanish,
	}

	defaultLanguage        = english
	defaultLanguageDetails = languageDetails[defaultLanguage]
)

func detailsForLanguage(l *displayLanguage) *languageDetail {
	switch l {
	case spanish:
		return spanishDetails
	case english:
		return englishDetails
	default:
		return defaultLanguageDetails
	}
}

func determineLanguage(req *http.Request) *displayLanguage {
	if req == nil {
		return defaultLanguage
	}

	langs, _, err := language.ParseAcceptLanguage(req.Header.Get("Accept-Language"))
	if err != nil {
		return defaultLanguage
	}

	if len(langs) != 1 {
		return defaultLanguage
	}

	switch langs[0] {
	case language.LatinAmericanSpanish, language.EuropeanSpanish, language.Spanish:
		return spanish
	case language.AmericanEnglish, language.BritishEnglish, language.English:
		return english
	default:
		return defaultLanguage
	}
}
