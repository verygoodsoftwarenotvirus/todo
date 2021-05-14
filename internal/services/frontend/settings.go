package frontend

import (
	// import embed for the side effect.
	_ "embed"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	observability "gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
)

//go:embed templates/partials/settings/user_settings.gotpl
var userSettingsPageSrc string

func (s *Service) buildUserSettingsView(includeBaseTemplate bool) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)
		tracing.AttachRequestToSpan(span, req)

		sessionCtxData, err := s.sessionContextDataFetcher(req)
		if err != nil {
			observability.AcknowledgeError(err, logger, span, "no session context data attached to request")
			http.Redirect(res, req, "/login", unauthorizedRedirectResponseCode)
			return
		}

		user, err := s.dataStore.GetUser(ctx, sessionCtxData.Requester.ID)
		if err != nil {
			observability.AcknowledgeError(err, logger, span, "fetching user from datastore")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		if includeBaseTemplate {
			tmpl := s.renderTemplateIntoBaseTemplate(userSettingsPageSrc, nil)

			page := &pageData{
				IsLoggedIn:  sessionCtxData != nil,
				Title:       "User Settings",
				ContentData: user,
			}
			if sessionCtxData != nil {
				page.IsServiceAdmin = sessionCtxData.Requester.ServiceAdminPermission.IsServiceAdmin()
			}

			s.renderTemplateToResponse(ctx, tmpl, page, res)
		} else {
			tmpl := s.parseTemplate(ctx, "", userSettingsPageSrc, nil)

			s.renderTemplateToResponse(ctx, tmpl, user, res)
		}
	}
}

//go:embed templates/partials/settings/account_settings.gotpl
var accountSettingsPageSrc string

func (s *Service) buildAccountSettingsView(includeBaseTemplate bool) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)
		tracing.AttachRequestToSpan(span, req)

		// get session context data
		sessionCtxData, err := s.sessionContextDataFetcher(req)
		if err != nil {
			observability.AcknowledgeError(err, logger, span, "no session context data attached to request")
			http.Redirect(res, req, buildRedirectURL("/login", "/account/settings"), unauthorizedRedirectResponseCode)
			return
		}

		account, err := s.fetchAccount(ctx, sessionCtxData)
		if err != nil {
			s.logger.Error(err, "retrieving account information from database")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		if includeBaseTemplate {
			tmpl := s.renderTemplateIntoBaseTemplate(accountSettingsPageSrc, nil)

			page := &pageData{
				IsLoggedIn:  sessionCtxData != nil,
				Title:       "Account Settings",
				ContentData: account,
			}
			if sessionCtxData != nil {
				page.IsServiceAdmin = sessionCtxData.Requester.ServiceAdminPermission.IsServiceAdmin()
			}

			s.renderTemplateToResponse(ctx, tmpl, page, res)
		} else {
			tmpl := s.parseTemplate(ctx, "", accountSettingsPageSrc, nil)

			s.renderTemplateToResponse(ctx, tmpl, account, res)
		}
	}
}

//go:embed templates/partials/settings/admin_settings.gotpl
var adminSettingsPageSrc string

func (s *Service) buildAdminSettingsView(includeBaseTemplate bool) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)

		// get session context data
		sessionCtxData, err := s.sessionContextDataFetcher(req)
		if err != nil {
			observability.AcknowledgeError(err, logger, span, "no session context data attached to request")
			http.Redirect(res, req, buildRedirectURL("/login", "/admin/settings"), unauthorizedRedirectResponseCode)
			return
		}

		if !sessionCtxData.Requester.ServiceAdminPermission.IsServiceAdmin() {
			observability.AcknowledgeError(err, logger, span, "no session context data attached to request")
			res.WriteHeader(http.StatusUnauthorized)
			return
		}

		if includeBaseTemplate {
			tmpl := s.renderTemplateIntoBaseTemplate(adminSettingsPageSrc, nil)

			page := &pageData{
				IsLoggedIn:  sessionCtxData != nil,
				Title:       "Admin Settings",
				ContentData: nil,
			}
			if sessionCtxData != nil {
				page.IsServiceAdmin = sessionCtxData.Requester.ServiceAdminPermission.IsServiceAdmin()
			}

			s.renderTemplateToResponse(ctx, tmpl, page, res)
		} else {
			tmpl := s.parseTemplate(ctx, "", adminSettingsPageSrc, nil)

			s.renderTemplateToResponse(ctx, tmpl, nil, res)
		}
	}
}
