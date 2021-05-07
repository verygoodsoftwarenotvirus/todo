package frontend

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
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

func (s *Service) buildUserSettingsView(includeBaseTemplate bool) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)

		sessionCtxData, err := s.sessionContextDataFetcher(req)
		if err != nil {
			observability.AcknowledgeError(err, logger, span, "no session context data attached to request")
			http.Redirect(res, req, "/login", http.StatusSeeOther)
			return
		}

		user, err := s.dataStore.GetUser(ctx, sessionCtxData.Requester.ID)
		if err != nil {
			observability.AcknowledgeError(err, logger, span, "error fetching user from datastore")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		if includeBaseTemplate {
			tmpl := s.renderTemplateIntoDashboard(userSettingsPageSrc, nil)

			page := &dashboardPageData{
				LoggedIn:    sessionCtxData != nil,
				Title:       "User Settings",
				ContentData: user,
			}

			if err = s.renderTemplateToResponse(tmpl, page, res); err != nil {
				observability.AcknowledgeError(err, logger, span, "rendering user settings viewer into dashboard")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			tmpl := s.parseTemplate("", userSettingsPageSrc, nil)

			if err = s.renderTemplateToResponse(tmpl, user, res); err != nil {
				observability.AcknowledgeError(err, logger, span, "rendering user settings viewer into dashboard")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}
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

func (s *Service) buildAccountSettingsView(includeBaseTemplate bool) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)

		// get session context data
		sessionCtxData, err := s.sessionContextDataFetcher(req)
		if err != nil {
			observability.AcknowledgeError(err, logger, span, "no session context data attached to request")
			http.Redirect(res, req, buildRedirectURL("/login", "/account/settings"), http.StatusSeeOther)
			return
		}

		account, err := s.fetchAccount(ctx, sessionCtxData)
		if err != nil {
			s.logger.Error(err, "retrieving account information from database")
			res.WriteHeader(http.StatusUnauthorized)
			return
		}

		if includeBaseTemplate {
			tmpl := s.renderTemplateIntoDashboard(accountSettingsPageSrc, nil)

			page := &dashboardPageData{
				LoggedIn:    sessionCtxData != nil,
				Title:       "Account Settings",
				ContentData: account,
			}

			if err = s.renderTemplateToResponse(tmpl, page, res); err != nil {
				observability.AcknowledgeError(err, logger, span, "rendering account settings viewer into dashboard")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			tmpl := s.parseTemplate("", accountSettingsPageSrc, nil)

			if err = s.renderTemplateToResponse(tmpl, account, res); err != nil {
				observability.AcknowledgeError(err, logger, span, "rendering account settings viewer into dashboard")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}
}
