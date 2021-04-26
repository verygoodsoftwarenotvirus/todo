package main

import (
	"log"
	"net/http"

	. "github.com/stevelacy/daz"
)

func homepage(w http.ResponseWriter, r *http.Request) {
	// <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@shoelace-style/shoelace@2.0.0-beta.39/dist/themes/base.css">
	shoelaceCSSImport := H(
		"link",
		Attr{
			"rel":  "stylesheet",
			"href": "https://cdn.jsdelivr.net/npm/@shoelace-style/shoelace@2.0.0-beta.39/dist/themes/base.css",
		},
	)
	//<script type="module" src="https://cdn.jsdelivr.net/npm/@shoelace-style/shoelace@2.0.0-beta.39/dist/shoelace.js"></script>
	shoelaceJSImport := H(
		"script",
		Attr{
			"type": "module",
			"src":  "https://cdn.jsdelivr.net/npm/@shoelace-style/shoelace@2.0.0-beta.39/dist/shoelace.js",
		},
	)

	//     <script src="https://unpkg.com/htmx.org@1.3.3" integrity="" crossorigin="anonymous"></script>
	htmxImport := H(
		"script",
		Attr{
			"src":         "https://unpkg.com/htmx.org@1.3.3",
			"integrity":   "sha384-QrlPmoLqMVfnV4lzjmvamY0Sv/Am8ca1W7veO++Sp6PiIGixqkD+0xZ955Nc03qO",
			"crossorigin": "anonymous",
		},
	)

	// <sl-button size="small">Click me</sl-button>
	exampleButton := H(
		"sl-button",
		Attr{
			"size":   "small",
			"hx-put": "/messages",
		},
		"Click me",
	)

	head := H("head", H("title", "HTMX+Shoelace Test Run"), shoelaceCSSImport, shoelaceJSImport, htmxImport)
	body := H("body", exampleButton)
	html := H("html", head, body)

	if _, err := w.Write([]byte(html())); err != nil {
		log.Fatalln(err)
	}
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", homepage)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalln(err)
	}
}
