package frontend

import (
	"context"
	_ "embed"
	"fmt"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"html/template"
	"io"
	"net/http"
	"net/url"
)

const (
	itemIDURLParamKey = "item"
)

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

//go:embed templates/partials/creators/item_creator.gotpl
var itemCreatorTemplate string

func (s *Service) buildItemCreatorView(includeBaseTemplate bool) func(http.ResponseWriter, *http.Request) {
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
			"componentTitle": func(x *types.Item) string {
				return fmt.Sprintf("Item #%d", x.ID)
			},
		}
		item := &types.Item{}

		if includeBaseTemplate {
			view := s.renderTemplateIntoBaseTemplate(itemCreatorTemplate, tmplFuncMap)

			page := &pageData{
				IsLoggedIn:  sessionCtxData != nil,
				Title:       "New Item",
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
			tmpl := s.parseTemplate("", itemCreatorTemplate, tmplFuncMap)

			if err = s.renderTemplateToResponse(tmpl, item, res); err != nil {
				observability.AcknowledgeError(err, logger, span, "rendering item editor view")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}
}

const (
	itemCreationInputNameFormKey    = "name"
	itemCreationInputDetailsFormKey = "details"
)

// parseFormEncodedItemCreationInput checks a request for an ItemCreationInput.
func (s *Service) parseFormEncodedItemCreationInput(ctx context.Context, req *http.Request, sessionCtxData *types.SessionContextData) (creationInput *types.ItemCreationInput) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "parsing item creation input")
		return nil
	}

	form, err := url.ParseQuery(string(bodyBytes))
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "parsing item creation input")
		return nil
	}

	creationInput = &types.ItemCreationInput{
		Name:             form.Get(itemCreationInputNameFormKey),
		Details:          form.Get(itemCreationInputDetailsFormKey),
		BelongsToAccount: sessionCtxData.ActiveAccountID,
	}

	if err = creationInput.ValidateWithContext(ctx); err != nil {
		logger = logger.WithValue("input", creationInput)
		observability.AcknowledgeError(err, logger, span, "invalid item creation input")
		return nil
	}

	return creationInput
}

func (s *Service) handleItemCreationRequest(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	logger.Debug("item Creation route called")

	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "no session context data attached to request")
		http.Redirect(res, req, "/login", http.StatusSeeOther)
		return
	}

	logger.Debug("session context data retrieved for item Creation route")

	creationInput := s.parseFormEncodedItemCreationInput(ctx, req, sessionCtxData)
	if creationInput == nil {
		observability.AcknowledgeError(err, logger, span, "parsing item creation input")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	logger.Debug("item Creation input parsed successfully")

	if _, err = s.dataStore.CreateItem(ctx, creationInput, sessionCtxData.Requester.ID); err != nil {
		observability.AcknowledgeError(err, logger, span, "writing item to datastore")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger.Debug("item Created")

	res.Header().Set(htmxRedirectionHeader, "/items")
	res.WriteHeader(http.StatusCreated)
}

//go:embed templates/partials/editors/item_editor.gotpl
var itemEditorTemplate string

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
			observability.AcknowledgeError(err, logger, span, "fetching item from datastore")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		logger = logger.WithValue(keys.ItemIDKey, item.ID)
		tmplFuncMap := map[string]interface{}{
			"componentTitle": func(x *types.Item) string {
				return fmt.Sprintf("Item #%d", x.ID)
			},
			"individualURL": func(x *types.Item) template.URL {
				return template.URL(fmt.Sprintf("/dashboard_pages/items/%d", x.ID))
			},
		}

		if includeBaseTemplate {
			view := s.renderTemplateIntoBaseTemplate(itemEditorTemplate, tmplFuncMap)

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
			tmpl := s.parseTemplate("", itemEditorTemplate, tmplFuncMap)

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

//go:embed templates/partials/tables/items_table.gotpl
var itemsTableTemplate string

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
			observability.AcknowledgeError(err, logger, span, "fetching items from datastore")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		tmplFuncMap := map[string]interface{}{
			"individualURL": func(x *types.Item) template.URL {
				return template.URL(fmt.Sprintf("/dashboard_pages/items/%d", x.ID))
			},
			"pushURL": func(x *types.Item) template.URL {
				return template.URL(fmt.Sprintf("/items/%d", x.ID))
			},
		}

		if includeBaseTemplate {
			tmpl := s.renderTemplateIntoBaseTemplate(itemsTableTemplate, tmplFuncMap)

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
			tmpl := s.parseTemplate("dashboard", itemsTableTemplate, tmplFuncMap)

			if err = s.renderTemplateToResponse(tmpl, items, res); err != nil {
				observability.AcknowledgeError(err, logger, span, "rendering items table view")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}
}

// parseFormEncodedItemUpdateInput checks a request for an ItemUpdateInput.
func (s *Service) parseFormEncodedItemUpdateInput(ctx context.Context, req *http.Request, sessionCtxData *types.SessionContextData) (creationInput *types.ItemUpdateInput) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "parsing item creation input")
		return nil
	}

	form, err := url.ParseQuery(string(bodyBytes))
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "parsing item creation input")
		return nil
	}

	creationInput = &types.ItemUpdateInput{
		Name:             form.Get(itemCreationInputNameFormKey),
		Details:          form.Get(itemCreationInputDetailsFormKey),
		BelongsToAccount: sessionCtxData.ActiveAccountID,
	}

	if err = creationInput.ValidateWithContext(ctx); err != nil {
		logger = logger.WithValue("input", creationInput)
		observability.AcknowledgeError(err, logger, span, "invalid item creation input")
		return nil
	}

	return creationInput
}

func (s *Service) handleItemUpdateRequest(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "no session context data attached to request")
		http.Redirect(res, req, "/login", http.StatusSeeOther)
		return
	}

	updateInput := s.parseFormEncodedItemUpdateInput(ctx, req, sessionCtxData)
	if updateInput == nil {
		observability.AcknowledgeError(err, logger, span, "no update input attached to request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	item, err := s.fetchItem(ctx, sessionCtxData, req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching item from datastore")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	changes := item.Update(updateInput)

	if err = s.dataStore.UpdateItem(ctx, item, sessionCtxData.Requester.ID, changes); err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching item from datastore")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger = logger.WithValue(keys.ItemIDKey, item.ID)
	tmplFuncMap := map[string]interface{}{
		"componentTitle": func(x *types.Item) string {
			return fmt.Sprintf("Item #%d", x.ID)
		},
		"individualURL": func(x *types.Item) template.URL {
			return template.URL(fmt.Sprintf("/dashboard_pages/items/%d", x.ID))
		},
	}

	tmpl := s.parseTemplate("", itemEditorTemplate, tmplFuncMap)

	if err = s.renderTemplateToResponse(tmpl, item, res); err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering item editor view")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Service) handleItemDeletionRequest(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "no session context data attached to request")
		http.Redirect(res, req, "/login", http.StatusSeeOther)
		return
	}

	itemID := s.routeParamManager.BuildRouteParamIDFetcher(logger, itemIDURLParamKey, "item")(req)
	if err = s.dataStore.ArchiveItem(ctx, itemID, sessionCtxData.Requester.ID, sessionCtxData.Requester.ID); err != nil {
		observability.AcknowledgeError(err, logger, span, "archiving items in datastore")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	items, err := s.fetchItems(ctx, sessionCtxData, req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching items from datastore")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	tmplFuncMap := map[string]interface{}{
		"individualURL": func(x *types.Item) template.URL {
			return template.URL(fmt.Sprintf("/dashboard_pages/items/%d", x.ID))
		},
		"pushURL": func(x *types.Item) template.URL {
			return template.URL(fmt.Sprintf("/items/%d", x.ID))
		},
	}

	tmpl := s.parseTemplate("dashboard", itemsTableTemplate, tmplFuncMap)

	if err = s.renderTemplateToResponse(tmpl, items, res); err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering items table view")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}
