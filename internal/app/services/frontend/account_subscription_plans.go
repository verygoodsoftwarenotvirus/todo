package frontend

import (
	"context"
	// import embed for the side effect.
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
)

const (
	accountSubscriptionPlanIDURLParamKey = "accountSubscriptionPlan"
)

func (s *Service) fetchAccountSubscriptionPlan(ctx context.Context, req *http.Request) (accountSubscriptionPlan *types.AccountSubscriptionPlan, err error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger
	accountSubscriptionPlanID := s.routeParamManager.BuildRouteParamIDFetcher(logger, accountSubscriptionPlanIDURLParamKey, "account subscription plan")(req)

	if s.useFakeData {
		accountSubscriptionPlan = fakes.BuildFakeAccountSubscriptionPlan()
	} else {
		accountSubscriptionPlan, err = s.dataStore.GetAccountSubscriptionPlan(ctx, accountSubscriptionPlanID)
		if err != nil {
			return nil, observability.PrepareError(err, logger, span, "fetching account subscription plan data")
		}
	}

	return accountSubscriptionPlan, nil
}

//go:embed templates/partials/generated/creators/account_subscription_plan_creator.gotpl
var accountSubscriptionPlanCreatorTemplate string

func (s *Service) buildAccountSubscriptionPlanCreatorView(includeBaseTemplate bool) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		_, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)

		sessionCtxData, err := s.sessionContextDataFetcher(req)
		if err != nil {
			observability.AcknowledgeError(err, logger, span, "no session context data attached to request")
			http.Redirect(res, req, "/login", http.StatusSeeOther)
			return
		}

		tmplFuncMap := map[string]interface{}{
			"componentTitle": func(x *types.AccountSubscriptionPlan) string {
				return fmt.Sprintf("AccountSubscriptionPlan #%d", x.ID)
			},
		}
		accountSubscriptionPlan := &types.AccountSubscriptionPlan{}

		if includeBaseTemplate {
			view := s.renderTemplateIntoBaseTemplate(accountSubscriptionPlanCreatorTemplate, tmplFuncMap)

			page := &pageData{
				IsLoggedIn:  sessionCtxData != nil,
				Title:       "New Account Subscription Plan",
				ContentData: accountSubscriptionPlan,
			}
			if sessionCtxData != nil {
				page.IsServiceAdmin = sessionCtxData.Requester.ServiceAdminPermission.IsServiceAdmin()
			}

			if err = s.renderTemplateToResponse(view, page, res); err != nil {
				observability.AcknowledgeError(err, logger, span, "rendering account subscription plans dashboard view")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			tmpl := s.parseTemplate("", accountSubscriptionPlanCreatorTemplate, tmplFuncMap)

			if err = s.renderTemplateToResponse(tmpl, accountSubscriptionPlan, res); err != nil {
				observability.AcknowledgeError(err, logger, span, "rendering account subscription plans editor view")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}
}

const (
	accountSubscriptionPlanCreationInputNameFormKey        = "name"
	accountSubscriptionPlanCreationInputDescriptionFormKey = "description"
	accountSubscriptionPlanCreationInputPriceFormKey       = "price"
	accountSubscriptionPlanCreationInputPeriodFormKey      = "period"
)

// parseFormEncodedAccountSubscriptionPlanCreationInput checks a request for an AccountSubscriptionPlanCreationInput.
func (s *Service) parseFormEncodedAccountSubscriptionPlanCreationInput(ctx context.Context, req *http.Request) (creationInput *types.AccountSubscriptionPlanCreationInput, err error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithRequest(req)

	form, err := s.extractFormFromRequest(ctx, req)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "parsing account subscription plans creation input")
	}

	rawPrice := form.Get(accountSubscriptionPlanCreationInputPriceFormKey)
	price, err := strconv.ParseUint(rawPrice, 10, 32)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "parsing account subscription plans price")
	}

	rawPeriod := form.Get(accountSubscriptionPlanCreationInputPeriodFormKey)
	period, err := time.ParseDuration(rawPeriod)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "parsing account subscription plans period")
	}

	creationInput = &types.AccountSubscriptionPlanCreationInput{
		Name:        form.Get(accountSubscriptionPlanCreationInputNameFormKey),
		Description: form.Get(accountSubscriptionPlanCreationInputDescriptionFormKey),
		Price:       uint32(price),
		Period:      period,
	}

	if err = creationInput.ValidateWithContext(ctx); err != nil {
		logger = logger.WithValue("input", creationInput)
		return nil, observability.PrepareError(err, logger, span, "invalid account subscription plans creation input")
	}

	return creationInput, nil
}

func (s *Service) handleAccountSubscriptionPlanCreationRequest(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	creationInput, err := s.parseFormEncodedAccountSubscriptionPlanCreationInput(ctx, req)
	if creationInput == nil || err != nil {
		observability.AcknowledgeError(err, logger, span, "parsing account subscription plans creation input")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	if _, err = s.dataStore.CreateAccountSubscriptionPlan(ctx, creationInput); err != nil {
		observability.AcknowledgeError(err, logger, span, "writing account subscription plans to datastore")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger.Debug("account subscription plans created")

	htmxRedirectTo(res, "/account_subscription_plans")
	res.WriteHeader(http.StatusCreated)
}

//go:embed templates/partials/generated/editors/account_subscription_plan_editor.gotpl
var accountSubscriptionPlanEditorTemplate string

func (s *Service) buildAccountSubscriptionPlanEditorView(includeBaseTemplate bool) func(http.ResponseWriter, *http.Request) {
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

		accountSubscriptionPlan, err := s.fetchAccountSubscriptionPlan(ctx, req)
		if err != nil {
			observability.AcknowledgeError(err, logger, span, "fetching account subscription plans from datastore")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		logger = logger.WithValue(keys.AccountSubscriptionPlanIDKey, accountSubscriptionPlan.ID)
		tmplFuncMap := map[string]interface{}{
			"componentTitle": func(x *types.AccountSubscriptionPlan) string {
				return fmt.Sprintf("AccountSubscriptionPlan #%d", x.ID)
			},
			"individualURL": func(x *types.AccountSubscriptionPlan) template.URL {
				/* #nosec G203 */
				return template.URL(fmt.Sprintf("/dashboard_pages/account_subscription_plans/%d", x.ID))
			},
		}

		if includeBaseTemplate {
			view := s.renderTemplateIntoBaseTemplate(accountSubscriptionPlanEditorTemplate, tmplFuncMap)

			page := &pageData{
				IsLoggedIn:  sessionCtxData != nil,
				Title:       fmt.Sprintf("Account Subscription Plan #%d", accountSubscriptionPlan.ID),
				ContentData: accountSubscriptionPlan,
			}
			if sessionCtxData != nil {
				page.IsServiceAdmin = sessionCtxData.Requester.ServiceAdminPermission.IsServiceAdmin()
			}

			if err = s.renderTemplateToResponse(view, page, res); err != nil {
				observability.AcknowledgeError(err, logger, span, "rendering account subscription plans dashboard view")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			tmpl := s.parseTemplate("", accountSubscriptionPlanEditorTemplate, tmplFuncMap)

			if err = s.renderTemplateToResponse(tmpl, accountSubscriptionPlan, res); err != nil {
				observability.AcknowledgeError(err, logger, span, "rendering account subscription plans editor view")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}
}

func (s *Service) fetchAccountSubscriptionPlans(ctx context.Context, req *http.Request) (accountSubscriptionPlans *types.AccountSubscriptionPlanList, err error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger

	if s.useFakeData {
		accountSubscriptionPlans = fakes.BuildFakeAccountSubscriptionPlanList()
	} else {
		filter := types.ExtractQueryFilter(req)
		// this can only be an admin request
		accountSubscriptionPlans, err = s.dataStore.GetAccountSubscriptionPlans(ctx, filter)

		if err != nil {
			return nil, observability.PrepareError(err, logger, span, "fetching accountSubscriptionPlan data")
		}
	}

	return accountSubscriptionPlans, nil
}

//go:embed templates/partials/generated/tables/account_subscription_plans_table.gotpl
var accountSubscriptionPlansTableTemplate string

func (s *Service) buildAccountSubscriptionPlansTableView(includeBaseTemplate bool) func(http.ResponseWriter, *http.Request) {
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

		accountSubscriptionPlans, err := s.fetchAccountSubscriptionPlans(ctx, req)
		if err != nil {
			observability.AcknowledgeError(err, logger, span, "fetching account subscription plans from datastore")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		tmplFuncMap := map[string]interface{}{
			"individualURL": func(x *types.AccountSubscriptionPlan) template.URL {
				/* #nosec G203 */
				return template.URL(fmt.Sprintf("/dashboard_pages/account_subscription_plans/%d", x.ID))
			},
			"pushURL": func(x *types.AccountSubscriptionPlan) template.URL {
				/* #nosec G203 */
				return template.URL(fmt.Sprintf("/accountSubscriptionPlans/%d", x.ID))
			},
		}

		if includeBaseTemplate {
			tmpl := s.renderTemplateIntoBaseTemplate(accountSubscriptionPlansTableTemplate, tmplFuncMap)

			page := &pageData{
				IsLoggedIn:  sessionCtxData != nil,
				Title:       "AccountSubscriptionPlans",
				ContentData: accountSubscriptionPlans,
			}
			if sessionCtxData != nil {
				page.IsServiceAdmin = sessionCtxData.Requester.ServiceAdminPermission.IsServiceAdmin()
			}

			if err = s.renderTemplateToResponse(tmpl, page, res); err != nil {
				observability.AcknowledgeError(err, logger, span, "rendering account subscription plans dashboard tmpl")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			tmpl := s.parseTemplate("dashboard", accountSubscriptionPlansTableTemplate, tmplFuncMap)

			if err = s.renderTemplateToResponse(tmpl, accountSubscriptionPlans, res); err != nil {
				observability.AcknowledgeError(err, logger, span, "rendering account subscription plans table view")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}
}

// parseFormEncodedAccountSubscriptionPlanUpdateInput checks a request for an AccountSubscriptionPlanUpdateInput.
func (s *Service) parseFormEncodedAccountSubscriptionPlanUpdateInput(ctx context.Context, req *http.Request) (updateInput *types.AccountSubscriptionPlanUpdateInput) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithRequest(req)

	form, err := s.extractFormFromRequest(ctx, req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "parsing account subscription plan creation input")
		return nil
	}

	rawPrice := form.Get(accountSubscriptionPlanCreationInputPriceFormKey)
	price, err := strconv.ParseUint(rawPrice, 10, 32)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "parsing account subscription plan price")
		return nil
	}

	rawPeriod := form.Get(accountSubscriptionPlanCreationInputPeriodFormKey)
	period, err := time.ParseDuration(rawPeriod)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "parsing account subscription plans period")
		return nil
	}

	updateInput = &types.AccountSubscriptionPlanUpdateInput{
		Name:        form.Get(accountSubscriptionPlanCreationInputNameFormKey),
		Description: form.Get(accountSubscriptionPlanCreationInputDescriptionFormKey),
		Price:       uint32(price),
		Period:      period,
	}

	if err = updateInput.ValidateWithContext(ctx); err != nil {
		logger = logger.WithValue("input", updateInput)
		observability.AcknowledgeError(err, logger, span, "invalid account subscription plans creation input")
		return nil
	}

	return updateInput
}

func (s *Service) handleAccountSubscriptionPlanUpdateRequest(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "no session context data attached to request")
		http.Redirect(res, req, "/login", http.StatusSeeOther)
		return
	}

	updateInput := s.parseFormEncodedAccountSubscriptionPlanUpdateInput(ctx, req)
	if updateInput == nil {
		observability.AcknowledgeError(err, logger, span, "no update input attached to request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	accountSubscriptionPlan, err := s.fetchAccountSubscriptionPlan(ctx, req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching account subscription plans from datastore")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	changes := accountSubscriptionPlan.Update(updateInput)

	if err = s.dataStore.UpdateAccountSubscriptionPlan(ctx, accountSubscriptionPlan, sessionCtxData.Requester.ID, changes); err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching account subscription plans from datastore")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger = logger.WithValue(keys.AccountSubscriptionPlanIDKey, accountSubscriptionPlan.ID)
	tmplFuncMap := map[string]interface{}{
		"componentTitle": func(x *types.AccountSubscriptionPlan) string {
			return fmt.Sprintf("Account Subscription Plan #%d", x.ID)
		},
		"individualURL": func(x *types.AccountSubscriptionPlan) template.URL {
			/* #nosec G203 */
			/* #nosec G203 */
			return template.URL(fmt.Sprintf("/dashboard_pages/account_subscription_plans/%d", x.ID))
		},
	}

	tmpl := s.parseTemplate("", accountSubscriptionPlanEditorTemplate, tmplFuncMap)

	if err = s.renderTemplateToResponse(tmpl, accountSubscriptionPlan, res); err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering account subscription plans editor view")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Service) handleAccountSubscriptionPlanDeletionRequest(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "no session context data attached to request")
		http.Redirect(res, req, "/login", http.StatusSeeOther)
		return
	}

	accountSubscriptionPlanID := s.routeParamManager.BuildRouteParamIDFetcher(logger, accountSubscriptionPlanIDURLParamKey, "accountSubscriptionPlan")(req)
	if err = s.dataStore.ArchiveAccountSubscriptionPlan(ctx, accountSubscriptionPlanID, sessionCtxData.Requester.ID); err != nil {
		observability.AcknowledgeError(err, logger, span, "archiving account subscription plans in datastore")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	accountSubscriptionPlans, err := s.fetchAccountSubscriptionPlans(ctx, req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching account subscription plans from datastore")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	tmplFuncMap := map[string]interface{}{
		"individualURL": func(x *types.AccountSubscriptionPlan) template.URL {
			/* #nosec G203 */
			/* #nosec G203 */
			return template.URL(fmt.Sprintf("/dashboard_pages/account_subscription_plans/%d", x.ID))
		},
		"pushURL": func(x *types.AccountSubscriptionPlan) template.URL {
			/* #nosec G203 */
			/* #nosec G203 */
			return template.URL(fmt.Sprintf("/accountSubscriptionPlans/%d", x.ID))
		},
	}

	tmpl := s.parseTemplate("dashboard", accountSubscriptionPlansTableTemplate, tmplFuncMap)

	if err = s.renderTemplateToResponse(tmpl, accountSubscriptionPlans, res); err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering account subscription plans table view")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}
