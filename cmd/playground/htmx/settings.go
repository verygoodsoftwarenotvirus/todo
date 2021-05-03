package main

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"html/template"
	"net/http"
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

var userSettingsComponentTemplate = template.Must(template.New("").Parse(userSettingsPageSrc))

func buildUserSettingsDashboardPage() string {
	u := fakes.BuildFakeUser()
	return buildDashboardSubpageString("User Settings", template.HTML(renderTemplateToString(userSettingsComponentTemplate, u)))
}

func userSettingsDashboardPage(res http.ResponseWriter, req *http.Request) {
	renderStringToResponse(buildUserSettingsDashboardPage())(res, req)
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

var accountSettingsComponentTemplate = template.Must(template.New("").Parse(accountSettingsPageSrc))

func buildAccountSettingsDashboardPage() string {
	a := fakes.BuildFakeAccount()
	return buildDashboardSubpageString("Account Settings", template.HTML(renderTemplateToString(accountSettingsComponentTemplate, a)))
}

func accountSettingsDashboardPage(res http.ResponseWriter, req *http.Request) {
	renderStringToResponse(buildAccountSettingsDashboardPage())(res, req)
}
