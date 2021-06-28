package frontend

import (
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
)

const (
	accountIDURLParamKey = "account"
)

func (s *service) fetchAccount(ctx context.Context, sessionCtxData *types.SessionContextData) (account *types.Account, err error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger

	if s.useFakeData {
		account = fakes.BuildFakeAccount()
	} else {
		account, err = s.dataStore.GetAccount(ctx, sessionCtxData.ActiveAccountID, sessionCtxData.Requester.UserID)
		if err != nil {
			return nil, observability.PrepareError(err, logger, span, "fetching account data")
		}
	}

	return account, nil
}

//go:embed templates/partials/generated/editors/account_editor.gotpl
var accountEditorTemplate string

func (s *service) buildAccountEditorView(includeBaseTemplate bool) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)
		tracing.AttachRequestToSpan(span, req)

		// get session context data
		sessionCtxData, err := s.sessionContextDataFetcher(req)
		if err != nil {
			observability.AcknowledgeError(err, logger, span, "no session context data attached to request")
			http.Redirect(res, req, "/login", unauthorizedRedirectResponseCode)
			return
		}

		account, err := s.fetchAccount(ctx, sessionCtxData)
		if err != nil {
			s.logger.Error(err, "retrieving account information from database")
			res.WriteHeader(http.StatusInternalServerError)
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
				page.IsServiceAdmin = sessionCtxData.Requester.ServicePermissions.IsServiceAdmin()
			}

			s.renderTemplateToResponse(ctx, tmpl, page, res)
		} else {
			tmpl := s.parseTemplate(ctx, "", accountEditorTemplate, templateFuncMap)

			s.renderTemplateToResponse(ctx, tmpl, account, res)
		}
	}
}

// plural

func (s *service) fetchAccounts(ctx context.Context, sessionCtxData *types.SessionContextData, req *http.Request) (accounts *types.AccountList, err error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger
	tracing.AttachRequestToSpan(span, req)

	if s.useFakeData {
		accounts = fakes.BuildFakeAccountList()
	} else {
		qf := types.ExtractQueryFilter(req)
		accounts, err = s.dataStore.GetAccounts(ctx, sessionCtxData.Requester.UserID, qf)
		if err != nil {
			return nil, observability.PrepareError(err, logger, span, "fetching accounts data")
		}
	}

	return accounts, nil
}

//go:embed templates/partials/generated/tables/accounts_table.gotpl
var accountsTableTemplate string

func (s *service) buildAccountsTableView(includeBaseTemplate bool) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)
		tracing.AttachRequestToSpan(span, req)

		// get session context data
		sessionCtxData, err := s.sessionContextDataFetcher(req)
		if err != nil {
			observability.AcknowledgeError(err, logger, span, "no session context data attached to request")
			http.Redirect(res, req, "/login", unauthorizedRedirectResponseCode)
			return
		}

		accounts, err := s.fetchAccounts(ctx, sessionCtxData, req)
		if err != nil {
			s.logger.Error(err, "retrieving account information from database")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		tmplFuncMap := map[string]interface{}{
			"individualURL": func(x *types.Account) template.URL {
				/* #nosec G203 */
				return template.URL(fmt.Sprintf("/accounts/%d", x.ID))
			},
			"pushURL": func(x *types.Account) template.URL {
				/* #nosec G203 */
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
				page.IsServiceAdmin = sessionCtxData.Requester.ServicePermissions.IsServiceAdmin()
			}

			s.renderTemplateToResponse(ctx, tmpl, page, res)
		} else {
			tmpl := s.parseTemplate(ctx, "dashboard", accountsTableTemplate, tmplFuncMap)

			s.renderTemplateToResponse(ctx, tmpl, accounts, res)
		}
	}
}
