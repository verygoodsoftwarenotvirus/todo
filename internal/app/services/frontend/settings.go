package frontend

import (
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

func buildUserSettingsDashboardPage() (string, error) {
	u := fakes.BuildFakeUser()

	return renderTemplateIntoDashboard("User Settings", userSettingsPageSrc, u)
}

func (s *Service) userSettingsDashboardPage(res http.ResponseWriter, _ *http.Request) {
	responseContent, err := buildUserSettingsDashboardPage()
	if err != nil {
		panic(err)
	}

	renderStringToResponse(responseContent, res)
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

func buildAccountSettingsDashboardPage() (string, error) {
	a := fakes.BuildFakeAccount()

	return renderTemplateIntoDashboard("Account Settings", accountSettingsPageSrc, a)
}

func (s *Service) accountSettingsDashboardPage(res http.ResponseWriter, req *http.Request) {
	responseContent, err := buildAccountSettingsDashboardPage()
	if err != nil {
		panic(err)
	}

	renderStringToResponse(responseContent, res)
}
