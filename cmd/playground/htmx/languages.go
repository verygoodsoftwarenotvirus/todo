package main

import (
	"net/http"

	"golang.org/x/text/language"
)

type (
	displayLanguage *string

	languageDetails struct {
		Name         string
		Abbreviation string
	}
)

var (
	english        = new(displayLanguage)
	englishDetails = languageDetails{
		Name:         "English",
		Abbreviation: "en-US",
	}

	spanish        = new(displayLanguage)
	spanishDetails = languageDetails{
		Name:         "Spanish",
		Abbreviation: "es-MX",
	}
)

func detailsForLanguage(l *displayLanguage) languageDetails {
	switch l {
	case spanish:
		return spanishDetails
	case english:
		return englishDetails
	default:
		return englishDetails
	}
}

func determineLanguage(req *http.Request) *displayLanguage {
	langs, _, err := language.ParseAcceptLanguage(req.Header.Get("Accept-Language"))
	if err != nil {
		return english
	}

	if len(langs) != 1 {
		return english
	}

	switch langs[0].String() {
	case "es-MX":
		return spanish
	case "en-US":
		return english
	default:
		return english
	}
}
