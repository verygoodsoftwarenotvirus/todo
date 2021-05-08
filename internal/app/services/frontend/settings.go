package frontend

import (
	// import embed for the side effect.
	_ "embed"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
)

//go:embed templates/user_settings_partial.gotpl
var userSettingsPageSrc string

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
			tmpl := s.renderTemplateIntoBaseTemplate(userSettingsPageSrc, nil)

			page := &pageData{
				IsLoggedIn:  sessionCtxData != nil,
				Title:       "User Settings",
				ContentData: user,
			}
			if sessionCtxData != nil {
				page.IsServiceAdmin = sessionCtxData.Requester.ServiceAdminPermission.IsServiceAdmin()
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

//go:embed templates/account_settings_partial.gotpl
var accountSettingsPageSrc string

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
			tmpl := s.renderTemplateIntoBaseTemplate(accountSettingsPageSrc, nil)

			page := &pageData{
				IsLoggedIn:  sessionCtxData != nil,
				Title:       "Account Settings",
				ContentData: account,
			}
			if sessionCtxData != nil {
				page.IsServiceAdmin = sessionCtxData.Requester.ServiceAdminPermission.IsServiceAdmin()
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

//go:embed templates/admin_settings_partial.gotpl
var adminSettingsPageSrc string

func (s *Service) buildAdminSettingsView(includeBaseTemplate bool) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		_, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)

		// get session context data
		sessionCtxData, err := s.sessionContextDataFetcher(req)
		if err != nil {
			observability.AcknowledgeError(err, logger, span, "no session context data attached to request")
			http.Redirect(res, req, buildRedirectURL("/login", "/admin/settings"), http.StatusSeeOther)
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

			if err = s.renderTemplateToResponse(tmpl, page, res); err != nil {
				observability.AcknowledgeError(err, logger, span, "rendering admin settings viewer into dashboard")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			tmpl := s.parseTemplate("", adminSettingsPageSrc, nil)

			if err = s.renderTemplateToResponse(tmpl, nil, res); err != nil {
				observability.AcknowledgeError(err, logger, span, "rendering admin settings viewer into dashboard")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}
}
