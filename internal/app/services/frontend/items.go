package frontend

import (
	"context"
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
)

const (
	itemIDURLParamKey = "item"
)

func (s *Service) buildItemsTableConfig() *basicTableTemplateConfig {
	return &basicTableTemplateConfig{
		Title:          "Items",
		ExternalURL:    "/items/123",
		CreatorPageURL: "/items/new",
		GetURL:         "/dashboard_pages/items/123",
		Columns:        s.fetchTableColumns("columns.items"),
		CellFields: []string{
			"Name",
			"Details",
		},
		RowDataFieldName:     "Items",
		IncludeLastUpdatedOn: true,
		IncludeCreatedOn:     true,
	}
}

func (s *Service) buildItemEditorConfig() *basicEditorTemplateConfig {
	return &basicEditorTemplateConfig{
		Fields: []basicEditorField{
			{
				Name:      "Name",
				InputType: "text",
				Required:  true,
			},
			{
				Name:      "Details",
				InputType: "text",
				Required:  false,
			},
		},
		FuncMap: map[string]interface{}{
			"componentTitle": func(x *types.Item) string {
				return fmt.Sprintf("Item #%d", x.ID)
			},
		},
	}
}

func (s *Service) fetchItem(ctx context.Context, sessionCtxData *types.SessionContextData, req *http.Request) (item *types.Item, err error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger
	itemID := s.routeParamManager.BuildRouteParamIDFetcher(logger, itemIDURLParamKey, "item")(req)

	if s.useFakeData {
		item = fakes.BuildFakeItem()
	} else {
		item, err = s.dataStore.GetItem(ctx, itemID, sessionCtxData.Requester.ID)
		if err != nil {
			return nil, observability.PrepareError(err, logger, span, "fetching item data")
		}
	}

	return item, nil
}

func (s *Service) buildItemEditorView(includeBaseTemplate bool) func(http.ResponseWriter, *http.Request) {
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

		item, err := s.fetchItem(ctx, sessionCtxData, req)
		if err != nil {
			observability.AcknowledgeError(err, logger, span, "error fetching item from datastore")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		logger = logger.WithValue(keys.ItemIDKey, item.ID)

		itemEditorConfig := s.buildItemEditorConfig()
		if includeBaseTemplate {
			view := s.renderTemplateIntoBaseTemplate(s.buildBasicEditorTemplate(itemEditorConfig), itemEditorConfig.FuncMap)

			page := &pageData{
				IsLoggedIn:  sessionCtxData != nil,
				Title:       fmt.Sprintf("Item #%d", item.ID),
				ContentData: item,
			}
			if sessionCtxData != nil {
				page.IsServiceAdmin = sessionCtxData.Requester.ServiceAdminPermission.IsServiceAdmin()
			}

			if err = s.renderTemplateToResponse(view, page, res); err != nil {
				observability.AcknowledgeError(err, logger, span, "rendering items dashboard view")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			tmpl := s.parseTemplate("", s.buildBasicEditorTemplate(itemEditorConfig), itemEditorConfig.FuncMap)

			if err = s.renderTemplateToResponse(tmpl, item, res); err != nil {
				observability.AcknowledgeError(err, logger, span, "rendering item editor view")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}
}

func (s *Service) fetchItems(ctx context.Context, sessionCtxData *types.SessionContextData, req *http.Request) (items *types.ItemList, err error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger

	if s.useFakeData {
		items = fakes.BuildFakeItemList()
	} else {
		filter := types.ExtractQueryFilter(req)
		if isAdminRequest(req) {
			items, err = s.dataStore.GetItemsForAdmin(ctx, filter)
		} else {
			items, err = s.dataStore.GetItems(ctx, sessionCtxData.Requester.ID, filter)
		}

		if err != nil {
			return nil, observability.PrepareError(err, logger, span, "fetching item data")
		}
	}

	return items, nil
}

func (s *Service) buildItemsTableView(includeBaseTemplate bool) func(http.ResponseWriter, *http.Request) {
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

		items, err := s.fetchItems(ctx, sessionCtxData, req)
		if err != nil {
			observability.AcknowledgeError(err, logger, span, "error fetching items from datastore")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		itemsTableConfig := s.buildItemsTableConfig()
		if includeBaseTemplate {
			tmpl := s.renderTemplateIntoBaseTemplate(s.buildBasicTableTemplate(itemsTableConfig), nil)

			page := &pageData{
				IsLoggedIn:  sessionCtxData != nil,
				Title:       "Items",
				ContentData: items,
			}
			if sessionCtxData != nil {
				page.IsServiceAdmin = sessionCtxData.Requester.ServiceAdminPermission.IsServiceAdmin()
			}

			if err = s.renderTemplateToResponse(tmpl, page, res); err != nil {
				observability.AcknowledgeError(err, logger, span, "rendering items dashboard tmpl")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			tmpl := s.parseTemplate("dashboard", s.buildBasicTableTemplate(itemsTableConfig), nil)

			if err = s.renderTemplateToResponse(tmpl, items, res); err != nil {
				observability.AcknowledgeError(err, logger, span, "rendering items table view")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}
}
