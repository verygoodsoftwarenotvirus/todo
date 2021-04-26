package main

import (
	"log"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	. "github.com/stevelacy/daz"
)

func exampleItemsTable(res http.ResponseWriter, r *http.Request) {
	exampleList := fakes.BuildFakeItemList()

	headers := []HTML{
		H("th", "ID"),
		H("th", "Name"),
		H("th", "Details"),
		H("th", "Created On"),
	}
	header := H("thead", H("tr", headers))

	var rows []HTML
	for _, i := range exampleList.Items {
		rows = append(rows,
			H("td", i.ID),
			H("td", i.Name),
			H("td", i.Details),
			H("td", i.CreatedOn),
		)
	}
	body := H("tr", rows)

	if _, err := res.Write([]byte(H("table", header, body)())); err != nil {
		log.Fatalln(err)
	}
}

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

	// <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/bulma/0.9.2/css/bulma.min.css" integrity="sha512-byErQdWdTqREz6DLAA9pCnLbdoGGhXfU6gm1c8bkf7F51JVmUBlayGe2A31VpXWQP+eiJ3ilTAZHCR3vmMyybA==" crossorigin="anonymous" />
	skeletonImport := H(
		"link",
		Attr{
			"src":         "https://cdnjs.cloudflare.com/ajax/libs/bulma/0.9.2/css/bulma.min.css",
			"integrity":   "sha512-byErQdWdTqREz6DLAA9pCnLbdoGGhXfU6gm1c8bkf7F51JVmUBlayGe2A31VpXWQP+eiJ3ilTAZHCR3vmMyybA==",
			"crossorigin": "anonymous",
			"rel":         "stylesheet",
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
			"size":      "small",
			"hx-get":    "/items",
			"hx-target": "#fart",
		},
		"Click me",
	)
	fart := H("div", Attr{"id": "fart"})

	head := H("head", H("title", "HTMX+Shoelace Test Run"), skeletonImport, shoelaceCSSImport, shoelaceJSImport, htmxImport)
	body := H("body", exampleButton, fart)
	html := H("html", head, body)

	if _, err := w.Write([]byte(html())); err != nil {
		log.Fatalln(err)
	}
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", homepage)
	mux.HandleFunc("/items", exampleItemsTable)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalln(err)
	}
}
