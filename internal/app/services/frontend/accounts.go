package frontend

import (
	"fmt"
	"log"
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

var accountEditorConfig = &basicEditorTemplateConfig{
	Fields: []basicEditorField{
		{
			Name:      "Name",
			InputType: "text",
			Required:  true,
		},
	},
	FuncMap: map[string]interface{}{
		"componentTitle": func(x *types.Account) string {
			return fmt.Sprintf("Account #%d", x.ID)
		},
	},
}

func (s *Service) accountsEditorView(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	var account *types.Account
	if s.useFakes {
		logger.Debug("using fakes")
		account = fakes.BuildFakeAccount()
	}

	tmpl := parseTemplate("", buildBasicEditorTemplate(accountEditorConfig), accountEditorConfig.FuncMap)

	if err := renderTemplateToResponse(tmpl, account, res); err != nil {
		log.Panic(err)
	}
}

func (s *Service) accountsTableView(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	var accounts *types.AccountList
	if s.useFakes {
		logger.Debug("using fakes")
		accounts = fakes.BuildFakeAccountList()
	}

	tmpl := parseTemplate("dashboard", buildBasicTableTemplate(accountsTableConfig), nil)

	if err := renderTemplateToResponse(tmpl, accounts, res); err != nil {
		log.Panic(err)
	}
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

	view := renderTemplateIntoDashboard(buildBasicEditorTemplate(accountEditorConfig), accountEditorConfig.FuncMap)

	page := &dashboardPageData{
		Title:       fmt.Sprintf("Account #%d", account.ID),
		ContentData: account,
	}

	if err := renderTemplateToResponse(view, page, res); err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering accounts dashboard view")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
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

	view := renderTemplateIntoDashboard(buildBasicTableTemplate(accountsTableConfig), nil)

	page := &dashboardPageData{
		Title:       "Accounts",
		ContentData: accounts,
	}

	if err := renderTemplateToResponse(view, page, res); err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering accounts dashboard view")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}
