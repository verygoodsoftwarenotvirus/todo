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

func homepage(res http.ResponseWriter, req *http.Request) {
	log.Println("homepage visisted")

	//// <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@shoelace-style/shoelace@2.0.0-beta.39/dist/themes/base.css">
	//shoelaceCSSImport := H(
	//	"link",
	//	Attr{
	//		"rel":  "stylesheet",
	//		"href": "https://cdn.jsdelivr.net/npm/@shoelace-style/shoelace@2.0.0-beta.39/dist/themes/base.css",
	//	},
	//)
	////<script type="module" src="https://cdn.jsdelivr.net/npm/@shoelace-style/shoelace@2.0.0-beta.39/dist/shoelace.js"></script>
	//shoelaceJSImport := H(
	//	"script",
	//	Attr{
	//		"type": "module",
	//		"src":  "https://cdn.jsdelivr.net/npm/@shoelace-style/shoelace@2.0.0-beta.39/dist/shoelace.js",
	//	},
	//)
	//
	//// <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.0.0-beta3/dist/css/bootstrap.min.css" rel="stylesheet" integrity="sha384-eOJMYsd53ii+scO/bJGFsiCZc+5NDVN2yr8+0RDqr0Ql0h+rP48ckxlpbzKgwra6" crossorigin="anonymous">
	//bootstrapImport := H(
	//	"link",
	//	Attr{
	//		"href":        "https://cdn.jsdelivr.net/npm/bootstrap@5.0.0-beta3/dist/css/bootstrap.min.css",
	//		"rel":         "stylesheet",
	//		"integrity":   "sha384-eOJMYsd53ii+scO/bJGFsiCZc+5NDVN2yr8+0RDqr0Ql0h+rP48ckxlpbzKgwra6",
	//		"crossorigin": "anonymous",
	//	},
	//)
	//
	////     <script src="https://unpkg.com/htmx.org@1.3.3" integrity="" crossorigin="anonymous"></script>
	//htmxImport := H(
	//	"script",
	//	Attr{
	//		"src":         "https://unpkg.com/htmx.org@1.3.3",
	//		"integrity":   "sha384-QrlPmoLqMVfnV4lzjmvamY0Sv/Am8ca1W7veO++Sp6PiIGixqkD+0xZ955Nc03qO",
	//		"crossorigin": "anonymous",
	//	},
	//)
	//
	//head := H("head", H("title", "HTMX+Shoelace Test Run"), bootstrapImport, shoelaceCSSImport, shoelaceJSImport, htmxImport)
	//body := H("body", htmlToInterface(buildLoginPrompt())...)
	//html := H("html", head, body)
	//
	//if _, err := res.Write([]byte(html())); err != nil {
	//	log.Fatalln(err)
	//}

	html := `
		<html>
			<head>
				<title>HTMX+Shoelace Test Run</title>
				<link href="https://cdn.jsdelivr.net/npm/bootstrap@5.0.0-beta3/dist/css/bootstrap.min.css" rel="stylesheet" integrity="sha384-eOJMYsd53ii+scO/bJGFsiCZc+5NDVN2yr8+0RDqr0Ql0h+rP48ckxlpbzKgwra6" crossorigin="anonymous">
				<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@shoelace-style/shoelace@2.0.0-beta.39/dist/themes/base.css">
				<script type="module" src="https://cdn.jsdelivr.net/npm/@shoelace-style/shoelace@2.0.0-beta.39/dist/shoelace.js"></script>
				<script src="https://unpkg.com/htmx.org@1.3.3" integrity="sha384-QrlPmoLqMVfnV4lzjmvamY0Sv/Am8ca1W7veO++Sp6PiIGixqkD+0xZ955Nc03qO" crossorigin="anonymous"></script>
				<style>.htmx-indicator{opacity:0;transition: opacity 200ms ease-in;} .htmx-request .htmx-indicator{opacity:1} .htmx-request.htmx-indicator{opacity:1}</style>
			</head>
			<body>
			    <header class="navbar navbar-dark sticky-top bg-dark flex-md-nowrap p-0 shadow">
			        <a class="navbar-brand col-md-3 col-lg-2 me-0 px-3" href="#">Company name</a>
			        <button class="navbar-toggler position-absolute d-md-none collapsed" type="button" data-bs-toggle="collapse" data-bs-target="#sidebarMenu" aria-controls="sidebarMenu" aria-expanded="false" aria-label="Toggle navigation">
			            <span class="navbar-toggler-icon"></span>
			        </button>
			        <ul class="navbar-nav px-3">
			            <li class="nav-item text-nowrap">
			                <a class="nav-link" href="#">Sign out</a>
			            </li>
			        </ul>
			    </header>
	
			    <div class="container-fluid">
			        <div class="row">
			            <nav id="sidebarMenu" class="col-md-3 col-lg-2 d-md-block bg-light sidebar collapse">
			                <div class="position-sticky pt-3">
			                    <ul class="nav flex-column">
			                        <li class="nav-item">
			                            <a class="nav-link active" aria-current="page" href="#">
			                                🏠 Dashboard
			                            </a>
			                        </li>
			                        <li class="nav-item">
			                            <a class="nav-link" href="#">
			                                💰 Customers
			                            </a>
			                        </li>
			                    </ul>
	
			                    <h6 class="sidebar-heading d-flex justify-content-between align-items-center px-3 mt-4 mb-1 text-muted">
			                        <span>Saved reports</span>
			                    </h6>
			                    <ul class="nav flex-column mb-2">
			                        <li class="nav-item">
			                            <a class="nav-link" href="#">
			                                📃 Current month
			                            </a>
			                        </li>
			                        <li class="nav-item">
			                            <a class="nav-link" href="#">
			                                📃 Last quarter
			                            </a>
			                        </li>
			                    </ul>
			                </div>
			            </nav>
	
			            <main class="col-md-9 ms-sm-auto col-lg-10 px-md-4">
			                <div class="chartjs-size-monitor">
			                    <div class="chartjs-size-monitor-expand"><div class=""></div></div>
			                    <div class="chartjs-size-monitor-shrink"><div class=""></div></div>
			                </div>
			                <div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
			                    <h1 class="h2">Dashboard</h1>
			                    <div class="btn-toolbar mb-2 mb-md-0">
			                        <div class="btn-group me-2">
			                            <button type="button" class="btn btn-sm btn-outline-secondary">Share</button>
			                            <button type="button" class="btn btn-sm btn-outline-secondary">Export</button>
			                        </div>
			                        <button type="button" class="btn btn-sm btn-outline-secondary dropdown-toggle">
			                            📅 This week
			                        </button>
			                    </div>
			                </div>
	
			                <p id="content"><button hx-get="/items" hx-target="#content">Get Items</button></p>
	
			            </main>
			        </div>
			    </div>
			</body>
		</html>
	`

	if _, err := res.Write([]byte(html)); err != nil {
		log.Fatalln(err)
	}
}

func buildLoginPrompt() []HTML {
	return []HTML{
		H(
			"form",
			Attr{
				"hx-post": "/login",
				//"hx-target": "#postLogin",
			},
			H(
				"h1",
				Attr{"class": "h3 mb-3 fw-normal"},
				"Sign in",
			),
			H(
				"div",
				Attr{"class": "form-floating"},
				H(
					"input",
					Attr{
						"type":        "text",
						"class":       "form-control",
						"id":          "usernameInput",
						"name":        "username",
						"value":       "username",
						"placeholder": "username",
					},
				),
				H(
					"label",
					Attr{"for": "usernameInput"},
					"Username",
				),
			),
			H(
				"div",
				Attr{"class": "form-floating"},
				H(
					"input",
					Attr{
						"type":        "text",
						"class":       "form-control",
						"id":          "passwordInput",
						"name":        "password",
						"value":       "password",
						"placeholder": "password",
					},
				),
				H(
					"label",
					Attr{"for": "passwordInput"},
					"Password",
				),
			),
			H(
				"div",
				Attr{"class": "form-floating"},
				H(
					"input",
					Attr{
						"type":        "text",
						"class":       "form-control",
						"id":          "totpTokenInput",
						"name":        "totpToken",
						"value":       "123456",
						"placeholder": "123456",
					},
				),
				H(
					"label",
					Attr{"for": "totpTokenInput"},
					"2FA Token",
				),
			),
			H(
				"button",
				Attr{
					"class": "w-100 btn btn-lg btn-primary",
					"type":  "submit",
				},
				"Sign in",
			),
		),
		//H("div", Attr{"id": "postLogin"}),
	}
}

func buildExampleButton() []HTML {
	return []HTML{
		H(
			"button",
			Attr{
				"hx-get": "/items",
			},
			"Press me!",
		),
	}
}

func buildDashboardShell() []HTML {
	/*
		<body>
		    <header class="navbar navbar-dark sticky-top bg-dark flex-md-nowrap p-0 shadow">
		        <a class="navbar-brand col-md-3 col-lg-2 me-0 px-3" href="#">Company name</a>
		        <button class="navbar-toggler position-absolute d-md-none collapsed" type="button" data-bs-toggle="collapse" data-bs-target="#sidebarMenu" aria-controls="sidebarMenu" aria-expanded="false" aria-label="Toggle navigation">
		            <span class="navbar-toggler-icon"></span>
		        </button>
		        <input class="form-control form-control-dark w-100" type="text" placeholder="Search" aria-label="Search" />
		        <ul class="navbar-nav px-3">
		            <li class="nav-item text-nowrap">
		                <a class="nav-link" href="#">Sign out</a>
		            </li>
		        </ul>
		    </header>

		    <div class="container-fluid">
		        <div class="row">
		            <nav id="sidebarMenu" class="col-md-3 col-lg-2 d-md-block bg-light sidebar collapse">
		                <div class="position-sticky pt-3">
		                    <ul class="nav flex-column">
		                        <li class="nav-item">
		                            <a class="nav-link active" aria-current="page" href="#">
		                                🏠 Dashboard
		                            </a>
		                        </li>
		                        <li class="nav-item">
		                            <a class="nav-link" href="#">
		                                💰 Customers
		                            </a>
		                        </li>
		                    </ul>

		                    <h6 class="sidebar-heading d-flex justify-content-between align-items-center px-3 mt-4 mb-1 text-muted">
		                        <span>Saved reports</span>
		                    </h6>
		                    <ul class="nav flex-column mb-2">
		                        <li class="nav-item">
		                            <a class="nav-link" href="#">
		                                📃 Current month
		                            </a>
		                        </li>
		                        <li class="nav-item">
		                            <a class="nav-link" href="#">
		                                📃 Last quarter
		                            </a>
		                        </li>
		                    </ul>
		                </div>
		            </nav>

		            <main class="col-md-9 ms-sm-auto col-lg-10 px-md-4">
		                <div class="chartjs-size-monitor">
		                    <div class="chartjs-size-monitor-expand"><div class=""></div></div>
		                    <div class="chartjs-size-monitor-shrink"><div class=""></div></div>
		                </div>
		                <div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
		                    <h1 class="h2">Dashboard</h1>
		                    <div class="btn-toolbar mb-2 mb-md-0">
		                        <div class="btn-group me-2">
		                            <button type="button" class="btn btn-sm btn-outline-secondary">Share</button>
		                            <button type="button" class="btn btn-sm btn-outline-secondary">Export</button>
		                        </div>
		                        <button type="button" class="btn btn-sm btn-outline-secondary dropdown-toggle">
		                            📅 This week
		                        </button>
		                    </div>
		                </div>

		                <p>I think content goes right here?</p>

		            </main>
		        </div>
		    </div>
		</body>
	*/
	return []HTML{
		//
	}
}

func htmlToInterface(in []HTML) []interface{} {
	x := []interface{}{}

	for _, y := range in {
		x = append(x, interface{}(y))
	}

	return x
}

func renderHTMLResponse(input []HTML) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		if _, err := res.Write([]byte(H("div", input)())); err != nil {
			log.Fatalln(err)
		}
	}
}

func divertTo(path string) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		http.Redirect(res, req, path, http.StatusFound)
	}
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", homepage)
	mux.HandleFunc("/login", divertTo("/items"))
	mux.HandleFunc("/items", exampleItemsTable)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalln(err)
	}
}
