package main

import (
	_html "html/template"
	"log"
	"net/http"
)

const rawDashboardTemplate = `{{ define "dashboard" }}
<html>
	<head>
		<title>TODO</title>
		<link href="https://cdn.jsdelivr.net/npm/bootstrap@5.0.0-beta3/dist/css/bootstrap.min.css" rel="stylesheet" integrity="sha384-eOJMYsd53ii+scO/bJGFsiCZc+5NDVN2yr8+0RDqr0Ql0h+rP48ckxlpbzKgwra6" crossorigin="anonymous">
		<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@shoelace-style/shoelace@2.0.0-beta.39/dist/themes/base.css">
		<script type="module" src="https://cdn.jsdelivr.net/npm/@shoelace-style/shoelace@2.0.0-beta.39/dist/shoelace.js"></script>
		<script src="https://unpkg.com/htmx.org@1.3.3" integrity="sha384-QrlPmoLqMVfnV4lzjmvamY0Sv/Am8ca1W7veO++Sp6PiIGixqkD+0xZ955Nc03qO" crossorigin="anonymous"></script>
	</head>
	<body>
		<header class="navbar navbar-dark sticky-top bg-dark flex-md-nowrap p-0 shadow">
			<a class="navbar-brand col-md-3 col-lg-2 me-0 px-3" href="#">TODO</a>
			<button class="navbar-toggler position-absolute d-md-none collapsed" type="button" data-bs-toggle="collapse" data-bs-target="#sidebarMenu" aria-controls="sidebarMenu" aria-expanded="false" aria-label="Toggle navigation">
				<span class="navbar-toggler-icon"></span>
			</button>
			<ul class="navbar-nav px-3">
				<li class="nav-item text-nowrap">
					<a class="nav-link" hx-target="#content" hx-get="/components/login_prompt">Sign in</a>
				</li>
			</ul>
		</header>

		<div class="container-fluid">
			<div class="row">
				<nav id="sidebarMenu" class="col-md-3 col-lg-2 d-md-block bg-light sidebar collapse">
					<div class="position-sticky pt-3">
						<h6 class="sidebar-heading d-flex justify-content-between align-items-center px-3 mt-4 mb-1 text-muted">
							<span>Things</span>
						</h6>
						<ul class="nav flex-column">
							<li class="nav-item">
								<a class="nav-link" hx-target="#content" hx-get="/dashboard_pages/items">
									📃 Items
								</a>
							</li>
							<li class="nav-item">
								<a class="nav-link"  aria-current="page" hx-target="#content" hx-get="/dashboard_pages/api_clients">
									🤖 API Clients
								</a>
							</li>
						</ul>

						<h6 class="sidebar-heading d-flex justify-content-between align-items-center px-3 mt-4 mb-1 text-muted">
							<span>Account</span>
						</h6>
						<ul class="nav flex-column">
							<li class="nav-item">
								<a class="nav-link"  aria-current="page" hx-target="#content" hx-get="/dashboard_pages/account/webhooks">
									🕸️ Webhooks
								</a>
							</li>
							<li class="nav-item">
								<a class="nav-link" hx-target="#content" hx-get="/dashboard_pages/account/settings">
									⚙️Settings
								</a>
							</li>
						</ul>

						<h6 class="sidebar-heading d-flex justify-content-between align-items-center px-3 mt-4 mb-1 text-muted">
							<span>User</span>
						</h6>
						<ul class="nav flex-column mb-2">
							<li class="nav-item">
								<a class="nav-link" hx-target="#content" hx-get="/dashboard_pages/accounts">
									📚 Accounts
								</a>
							</li>
						</ul>
						<ul class="nav flex-column mb-2">
							<li class="nav-item">
								<a class="nav-link" hx-target="#content" hx-get="/dashboard_pages/user_settings">
									⚙️Settings
								</a>
							</li>
						</ul>
					</div>
				</nav>

				<main class="col-md-9 ms-sm-auto col-lg-10 px-md-4">
					<div id="content">
						{{ .RawHTML }}
					</div>
				</main>
			</div>
		</div>
	</body>
</html>
{{ end }}
`

var dashboardTemplate = _html.Must(_html.New("dashboard").Parse(rawDashboardTemplate))

type dashboardPage struct {
	Title   string
	RawHTML _html.HTML
}

func renderRawStringIntoDashboard(thing string) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, _ *http.Request) {
		log.Println("dashboard visited")

		x := &dashboardPage{
			Title:   "Dashboard",
			RawHTML: _html.HTML(thing),
		}

		if err := dashboardTemplate.Execute(res, x); err != nil {
			log.Fatalln(err)
		}
	}
}
