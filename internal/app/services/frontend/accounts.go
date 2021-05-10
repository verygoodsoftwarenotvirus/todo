package frontend

import (
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
)

const (
	accountIDURLParamKey = "account"
)

func (s *Service) fetchAccount(ctx context.Context, sessionCtxData *types.SessionContextData) (account *types.Account, err error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger

	if s.useFakeData {
		account = fakes.BuildFakeAccount()
	} else {
		account, err = s.dataStore.GetAccount(ctx, sessionCtxData.ActiveAccountID, sessionCtxData.Requester.ID)
		if err != nil {
			return nil, observability.PrepareError(err, logger, span, "fetching account data")
		}
	}

	return account, nil
}

//go:embed templates/partials/editors/account_editor.gotpl
var accountEditorTemplate string

func (s *Service) buildAccountView(includeBaseTemplate bool) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)

		// get session context data
		sessionCtxData, err := s.sessionContextDataFetcher(req)
		if err != nil {
			observability.AcknowledgeError(err, logger, span, "no session context data attached to request")
			http.Redirect(res, req, "/login", http.StatusSeeOther)
			return
		}

		account, err := s.fetchAccount(ctx, sessionCtxData)
		if err != nil {
			s.logger.Error(err, "retrieving account information from database")
			res.WriteHeader(http.StatusUnauthorized)
			return
		}

		templateFuncMap := map[string]interface{}{
			"componentTitle": func(x *types.Account) string {
				return fmt.Sprintf("Account #%d", x.ID)
			},
		}

		if includeBaseTemplate {
			tmpl := s.renderTemplateIntoBaseTemplate(accountEditorTemplate, templateFuncMap)

			page := &pageData{
				IsLoggedIn:  sessionCtxData != nil,
				Title:       fmt.Sprintf("Account #%d", account.ID),
				ContentData: account,
			}
			if sessionCtxData != nil {
				page.IsServiceAdmin = sessionCtxData.Requester.ServiceAdminPermission.IsServiceAdmin()
			}

			if err = s.renderTemplateToResponse(tmpl, page, res); err != nil {
				observability.AcknowledgeError(err, logger, span, "rendering accounts dashboard tmpl")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			tmpl := s.parseTemplate("", accountEditorTemplate, templateFuncMap)

			if err = s.renderTemplateToResponse(tmpl, account, res); err != nil {
				observability.AcknowledgeError(err, logger, span, "rendering accounts editor view")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}
}

// plural

func (s *Service) fetchAccounts(ctx context.Context, req *http.Request, sessionCtxData *types.SessionContextData) (accounts *types.AccountList, err error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger

	if s.useFakeData {
		accounts = fakes.BuildFakeAccountList()
	} else {
		qf := types.ExtractQueryFilter(req)
		accounts, err = s.dataStore.GetAccounts(ctx, sessionCtxData.Requester.ID, qf)
		if err != nil {
			return nil, observability.PrepareError(err, logger, span, "fetching accounts data")
		}
	}

	return accounts, nil
}

//go:embed templates/partials/tables/accounts_table.gotpl
var accountsTableTemplate string

func (s *Service) buildAccountsView(includeBaseTemplate bool) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)

		// get session context data
		sessionCtxData, err := s.sessionContextDataFetcher(req)
		if err != nil {
			observability.AcknowledgeError(err, logger, span, "no session context data attached to request")
			http.Redirect(res, req, "/login", http.StatusSeeOther)
			return
		}

		accounts, err := s.fetchAccounts(ctx, req, sessionCtxData)
		if err != nil {
			s.logger.Error(err, "retrieving account information from database")
			res.WriteHeader(http.StatusUnauthorized)
			return
		}

		tmplFuncMap := map[string]interface{}{
			"individualURL": func(x *types.Account) template.URL {
				return template.URL(fmt.Sprintf("/accounts/%d", x.ID))
			},
			"pushURL": func(x *types.Account) template.URL {
				return template.URL(fmt.Sprintf("/accounts/%d", x.ID))
			},
		}

		if includeBaseTemplate {
			tmpl := s.renderTemplateIntoBaseTemplate(accountsTableTemplate, tmplFuncMap)

			page := &pageData{
				IsLoggedIn:  sessionCtxData != nil,
				Title:       "Accounts",
				ContentData: accounts,
			}
			if sessionCtxData != nil {
				page.IsServiceAdmin = sessionCtxData.Requester.ServiceAdminPermission.IsServiceAdmin()
			}

			if err = s.renderTemplateToResponse(tmpl, page, res); err != nil {
				observability.AcknowledgeError(err, logger, span, "rendering accounts dashboard tmpl")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			tmpl := s.parseTemplate("dashboard", accountsTableTemplate, tmplFuncMap)

			if err = s.renderTemplateToResponse(tmpl, accounts, res); err != nil {
				observability.AcknowledgeError(err, logger, span, "rendering accounts dashboard view")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}
}
