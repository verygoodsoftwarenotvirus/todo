package elements

import (
	"html/template"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
)

const userSettingsPageSrc = `<div class="col-md-8 order-md-1">
	<form class="needs-validation" novalidate="">
		<div class="mb3">
			<label for="Username">Username</label>
			<div class="input-group">
				<input class="form-control" type="text" id="Username" placeholder="Username"required="" value="{{ .Username }}" />
				<div class="invalid-feedback" style="width: 100%;">Name is required.</div>
			</div>
		</div>
		
		<hr class="mb-4" />
		<button class="btn btn-primary btn-lg btn-block" type="submit">Save</button>
	</form>
</div>`

func buildUserSettingsDashboardPage() template.HTML {
	u := fakes.BuildFakeUser()
	return buildDashboardSubpageString("User Settings", renderTemplateToHTML(template.Must(template.New("").Parse(userSettingsPageSrc)), u))
}

func userSettingsDashboardPage(res http.ResponseWriter, req *http.Request) {
	renderHTMLTemplateToResponse(buildUserSettingsDashboardPage())(res, req)
}

const accountSettingsPageSrc = `<div class="col-md-8 order-md-1">
	<form class="needs-validation" novalidate="">
		<div class="mb3">
			<label for="Name">Name</label>
			<div class="input-group">
				<input class="form-control" type="text" id="Name" placeholder="Name"required="" value="{{ .Name }}" />
				<div class="invalid-feedback" style="width: 100%;">Name is required.</div>
			</div>
		</div>
		
		<hr class="mb-4" />
		<button class="btn btn-primary btn-lg btn-block" type="submit">Save</button>
	</form>
</div>`

func buildAccountSettingsDashboardPage() template.HTML {
	a := fakes.BuildFakeAccount()
	return buildDashboardSubpageString("Account Settings", renderTemplateToHTML(template.Must(template.New("").Parse(accountSettingsPageSrc)), a))
}

func accountSettingsDashboardPage(res http.ResponseWriter, req *http.Request) {
	renderHTMLTemplateToResponse(buildAccountSettingsDashboardPage())(res, req)
}
