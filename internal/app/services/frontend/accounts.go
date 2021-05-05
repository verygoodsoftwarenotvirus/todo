package frontend

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
)

var accountsTableConfig = &basicTableTemplateConfig{
	Title:       "Accounts",
	ExternalURL: "/accounts/123",
	GetURL:      "/dashboard_pages/accounts/123",
	Columns:     fetchTableColumns("columns.accounts"),
	CellFields: []string{
		"Name",
		"ExternalID",
		"BelongsToUser",
	},
	RowDataFieldName:     "Accounts",
	IncludeLastUpdatedOn: true,
	IncludeCreatedOn:     true,
}

func buildViewerForAccount(x *types.Account) (string, error) {
	tmplConfig := &basicEditorTemplateConfig{
		Name: "Account",
		ID:   x.ID,
		Fields: []genericEditorField{
			{
				Name:      "Name",
				InputType: "text",
				Required:  true,
			},
		},
	}

	tmpl := parseTemplate("", buildBasicEditorTemplate(tmplConfig))

	return renderTemplateToString(tmpl, x)
}

func (s *Service) accountDashboardPage(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	var account *types.Account
	if s.useFakes {
		logger.Debug("using fakes")
		account = fakes.BuildFakeAccount()
	}

	page, err := buildViewerForAccount(account)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering account table template into dashboard")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	renderStringToResponse(page, res)
}

func buildAccountsTableDashboardPage(accounts *types.AccountList) (string, error) {
	tmpl := parseTemplate("dashboard", buildBasicTableTemplate(accountsTableConfig))

	return renderTemplateToString(tmpl, accounts)
}

func (s *Service) accountsDashboardPage(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	var accounts *types.AccountList
	if s.useFakes {
		logger.Debug("using fakes")
		accounts = fakes.BuildFakeAccountList()
	}

	page, err := buildAccountsTableDashboardPage(accounts)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering account table template into dashboard")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	renderStringToResponse(page, res)
}

func buildAccountDashboardView(x *types.Account) (string, error) {
	tmplConfig := &basicEditorTemplateConfig{
		Name: "Account",
		ID:   x.ID,
		Fields: []genericEditorField{
			{
				Name:      "Name",
				InputType: "text",
				Required:  true,
			},
		},
	}

	return renderTemplateIntoDashboard("Accounts", wrapTemplateInContentDefinition(buildBasicEditorTemplate(tmplConfig)), x)
}

func (s *Service) accountDashboardView(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	var account *types.Account
	if s.useFakes {
		logger.Debug("using fakes")
		account = fakes.BuildFakeAccount()
	}

	dashboard, err := buildAccountDashboardView(account)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering account viewer into dashboard")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	renderStringToResponse(dashboard, res)
}

func buildAccountsTableDashboardView(accounts *types.AccountList) (string, error) {
	tmpl := wrapTemplateInContentDefinition(buildBasicTableTemplate(accountsTableConfig))

	return renderTemplateIntoDashboard("Accounts", tmpl, accounts)
}

func (s *Service) accountsDashboardView(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	var accounts *types.AccountList
	if s.useFakes {
		logger.Debug("using fakes")
		accounts = fakes.BuildFakeAccountList()
	}

	dashboard, err := buildAccountsTableDashboardView(accounts)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering account table template into dashboard")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	renderStringToResponse(dashboard, res)
}
