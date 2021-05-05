package frontend

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
)

const userSettingsPageSrc = `<div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
    <h1 class="h2">User Settings</h1>
</div>
<div class="col-md-8 order-md-1">
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

func (s *Service) userSettingsDashboardPage(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	var u *types.User
	if s.useFakes {
		u = fakes.BuildFakeUser()
	}

	responseContent, err := renderTemplateToString(parseTemplate("", userSettingsPageSrc), u)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering item viewer into dashboard")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	renderStringToResponse(responseContent, res)
}

func (s *Service) userSettingsDashboardView(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	var u *types.User
	if s.useFakes {
		u = fakes.BuildFakeUser()
	}

	responseContent, err := renderTemplateIntoDashboard("User Settings", wrapTemplateInContentDefinition(userSettingsPageSrc), u)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering item viewer into dashboard")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	renderStringToResponse(responseContent, res)
}

const accountSettingsPageSrc = `<div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
    <h1 class="h2">Account Settings</h1>
</div>
<div class="col-md-8 order-md-1">
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

func (s *Service) accountSettingsDashboardPage(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	var account *types.Account
	if s.useFakes {
		account = fakes.BuildFakeAccount()
	}

	responseContent, err := renderTemplateToString(parseTemplate("", accountSettingsPageSrc), account)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering item viewer into dashboard")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	renderStringToResponse(responseContent, res)
}

func (s *Service) accountSettingsDashboardView(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	var a *types.Account
	if s.useFakes {
		a = fakes.BuildFakeAccount()
	}

	responseContent, err := renderTemplateIntoDashboard("Account Settings", wrapTemplateInContentDefinition(accountSettingsPageSrc), a)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering item viewer into dashboard")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	renderStringToResponse(responseContent, res)
}
