package main

import (
	"html/template"
	"log"
	"net/http"
)

const dashboardTemplateSrc = `<!DOCTYPE html>
<html lang="en" class="no-js">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width">

		<title>TODO</title>

		<script type="module">
			document.documentElement.classList.remove('no-js');
			document.documentElement.classList.add('js');
		</script>

		<link href="https://cdn.jsdelivr.net/npm/bootstrap@5.0.0-beta3/dist/css/bootstrap.min.css" rel="stylesheet" integrity="sha384-eOJMYsd53ii+scO/bJGFsiCZc+5NDVN2yr8+0RDqr0Ql0h+rP48ckxlpbzKgwra6" crossorigin="anonymous">

		<!-- <meta name="description" content="{{ "{{ .PageDescription }}" }}">                     -->
		<!-- <meta property="og:title" content="{{ "{{ .PageTitle }}" }}">         -->
		<!-- <meta property="og:description" content="{{ "{{ .PageDescription }}" }}">              -->
		<!-- {{ if ne .PageImagePreview "" }}<meta property="og:image" content="{{ "{{ .PageImagePreview }}" }}">{{ end }} -->
		<!-- {{ if and (ne .PageImagePreview "") (ne .PageImagePreviewDescription "") }}<meta property="og:image:alt" content="{{ .PageImagePreviewDescription }}">{{ end }} -->  
		<!-- <meta property="og:locale" content="en_GB">                              -->
		<!-- <meta property="og:type" content="website">                              -->
		<!-- <meta name="twitter:card" content="summary_large_image">                 -->
		<!-- <meta property="og:url" content="https://www.mywebsite.com/page">        -->
		<!-- <link rel="canonical" href="https://www.mywebsite.com/page">             -->

		<link rel="icon" href="/favicon.ico">
		<link rel="icon" href="/favicon.svg" type="image/svg+xml">
		<link rel="apple-touch-icon" href="/apple-touch-icon.png">
		<!-- <link rel="manifest" href="/my.webmanifest">                             -->
		<!-- <meta name="theme-color" content="#FF00FF">                              -->
	</head>
	<body>
		<header class="navbar navbar-dark sticky-top bg-dark flex-md-nowrap p-0 shadow">
			<a class="navbar-brand col-md-3 col-lg-2 me-0 px-3" href="#">TODO</a>
			<button class="navbar-toggler position-absolute d-md-none collapsed" type="button" data-bs-toggle="collapse" data-bs-target="#sidebarMenu" aria-controls="sidebarMenu" aria-expanded="false" aria-label="Toggle navigation">
				<span class="navbar-toggler-icon"></span>
			</button>
			<ul class="navbar-nav px-3">
				<li class="nav-item text-nowrap">
					<a class="nav-link" hx-target="#content" hx-get="/components/login_prompt">{{ translate "callsToAction.signIn" }}</a>
				</li>
			</ul>
		</header>

		<div class="container-fluid">
			<div class="row">
				<nav id="sidebarMenu" class="col-md-3 col-lg-2 d-md-block bg-light sidebar collapse">
					<div class="position-sticky pt-3">
						<h6 class="sidebar-heading d-flex justify-content-between align-items-center px-3 mt-4 mb-1 text-muted">
							<span>{{ translate "dashboardSidebar.labels.things" }}</span>
						</h6>
						<ul class="nav flex-column">
							<li class="nav-item">
								<a class="nav-link" hx-target="#content" hx-push-url="/items" hx-get="/dashboard_pages/items">
									📃 {{ translate "dashboardSidebar.pageLinks.items" }}
								</a>
							</li>
							<li class="nav-item">
								<a class="nav-link"  aria-current="page" hx-target="#content" hx-push-url="/api_clients" hx-get="/dashboard_pages/api_clients">
									🤖 {{ translate "dashboardSidebar.pageLinks.apiClients" }}
								</a>
							</li>
						</ul>

						<h6 class="sidebar-heading d-flex justify-content-between align-items-center px-3 mt-4 mb-1 text-muted">
							<span>{{ translate "dashboardSidebar.labels.account" }}</span>
						</h6>
						<ul class="nav flex-column">
							<li class="nav-item">
								<a class="nav-link"  aria-current="page" hx-target="#content" hx-push-url="/account/webhooks" hx-get="/dashboard_pages/account/webhooks">
									🕸️ {{ translate "dashboardSidebar.pageLinks.webhooks" }}
								</a>
							</li>
							<li class="nav-item">
								<a class="nav-link" hx-target="#content" hx-push-url="/account/settings" hx-get="/dashboard_pages/account/settings">
									⚙ {{ translate "dashboardSidebar.pageLinks.accountSettings" }}
								</a>
							</li>
						</ul>

						<h6 class="sidebar-heading d-flex justify-content-between align-items-center px-3 mt-4 mb-1 text-muted">
							<span>{{ translate "dashboardSidebar.labels.user" }}</span>
						</h6>
						<ul class="nav flex-column mb-2">
							<li class="nav-item">
								<a class="nav-link" hx-target="#content" hx-push-url="/accounts" hx-get="/dashboard_pages/accounts">
									📚 {{ translate "dashboardSidebar.pageLinks.accounts" }}
								</a>
							</li>
						</ul>
						<ul class="nav flex-column mb-2">
							<li class="nav-item">
								<a class="nav-link" hx-target="#content" hx-push-url="/user/settings" hx-get="/dashboard_pages/user/settings">
									⚙{{ translate "dashboardSidebar.pageLinks.userSettings" }}
								</a>
							</li>
						</ul>
					</div>
				</nav>

				<main class="col-md-9 ms-sm-auto col-lg-10 px-md-4">
					<div id="content">
						{{ .Page }}
					</div>
				</main>
			</div>
		</div>

		<script src="https://unpkg.com/htmx.org@1.3.3" integrity="sha384-QrlPmoLqMVfnV4lzjmvamY0Sv/Am8ca1W7veO++Sp6PiIGixqkD+0xZ955Nc03qO" crossorigin="anonymous"></script>
	</body>
</html>
`

var dashboardTemplate = template.Must(template.New("").Funcs(defaultFuncMap).Parse(dashboardTemplateSrc))

type dashboardPageData struct {
	Title                       string
	Page                        template.HTML
	PageDescription             string
	PageTitle                   string
	PageImagePreview            string
	PageImagePreviewDescription string
}

func renderRawStringIntoDashboard(thing string) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, _ *http.Request) {
		x := &dashboardPageData{
			Title: "Dashboard",
			Page:  template.HTML(thing),
		}

		if err := dashboardTemplate.Execute(res, x); err != nil {
			log.Fatalln(err)
		}
	}
}

const dashboardPageTemplateFormat = `<div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
	<h1 class="h2">{{ .Title }}</h1>
</div>
{{ .Page }}
`

var dashboardPageTemplate = template.Must(template.New("").Parse(dashboardPageTemplateFormat))

func buildDashboardSubpageString(title string, content template.HTML) string {
	x := &dashboardPageData{
		Page:  content,
		Title: title,
	}
	return renderTemplateToString(dashboardPageTemplate, x)
}
