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
	itemIDURLParamKey = "item"
)

func (s *service) fetchItem(ctx context.Context, sessionCtxData *types.SessionContextData, req *http.Request) (item *types.Item, err error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger
	tracing.AttachRequestToSpan(span, req)

	if s.useFakeData {
		item = fakes.BuildFakeItem()
	} else {
		itemID := s.routeParamManager.BuildRouteParamIDFetcher(logger, itemIDURLParamKey, "item")(req)
		item, err = s.dataStore.GetItem(ctx, itemID, sessionCtxData.ActiveAccountID)
		if err != nil {
			return nil, observability.PrepareError(err, logger, span, "fetching item data")
		}
	}

	return item, nil
}

//go:embed templates/partials/generated/creators/item_creator.gotpl
var itemCreatorTemplate string

func (s *service) buildItemCreatorView(includeBaseTemplate bool) func(http.ResponseWriter, *http.Request) {
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

		item := &types.Item{}
		if includeBaseTemplate {
			view := s.renderTemplateIntoBaseTemplate(itemCreatorTemplate, nil)

			page := &pageData{
				IsLoggedIn:  sessionCtxData != nil,
				Title:       "New Item",
				ContentData: item,
			}
			if sessionCtxData != nil {
				page.IsServiceAdmin = sessionCtxData.Requester.ServicePermissions.IsServiceAdmin()
			}

			s.renderTemplateToResponse(ctx, view, page, res)
		} else {
			tmpl := s.parseTemplate(ctx, "", itemCreatorTemplate, nil)

			s.renderTemplateToResponse(ctx, tmpl, item, res)
		}
	}
}

const (
	nameFormKey    = "name"
	detailsFormKey = "details"

	itemCreationInputNameFormKey    = nameFormKey
	itemCreationInputDetailsFormKey = detailsFormKey

	itemUpdateInputNameFormKey    = nameFormKey
	itemUpdateInputDetailsFormKey = detailsFormKey
)

// parseFormEncodedItemCreationInput checks a request for an ItemCreationInput.
func (s *service) parseFormEncodedItemCreationInput(ctx context.Context, req *http.Request, sessionCtxData *types.SessionContextData) (creationInput *types.ItemCreationInput) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	form, err := s.extractFormFromRequest(ctx, req)
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

func (s *service) handleItemCreationRequest(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	logger.Debug("item creation route called")

	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "no session context data attached to request")
		http.Redirect(res, req, "/login", unauthorizedRedirectResponseCode)
		return
	}

	logger.Debug("session context data retrieved for item creation route")

	creationInput := s.parseFormEncodedItemCreationInput(ctx, req, sessionCtxData)
	if creationInput == nil {
		observability.AcknowledgeError(err, logger, span, "parsing item creation input")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	logger.Debug("item creation input parsed successfully")

	if _, err = s.dataStore.CreateItem(ctx, creationInput, sessionCtxData.Requester.UserID); err != nil {
		observability.AcknowledgeError(err, logger, span, "writing item to datastore")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger.Debug("item created")

	htmxRedirectTo(res, "/items")
	res.WriteHeader(http.StatusCreated)
}

//go:embed templates/partials/generated/editors/item_editor.gotpl
var itemEditorTemplate string

func (s *service) buildItemEditorView(includeBaseTemplate bool) func(http.ResponseWriter, *http.Request) {
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

		item, err := s.fetchItem(ctx, sessionCtxData, req)
		if err != nil {
			observability.AcknowledgeError(err, logger, span, "fetching item from datastore")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		tmplFuncMap := map[string]interface{}{
			"componentTitle": func(x *types.Item) string {
				return fmt.Sprintf("Item #%d", x.ID)
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
				page.IsServiceAdmin = sessionCtxData.Requester.ServicePermissions.IsServiceAdmin()
			}

			s.renderTemplateToResponse(ctx, view, page, res)
		} else {
			tmpl := s.parseTemplate(ctx, "", itemEditorTemplate, tmplFuncMap)

			s.renderTemplateToResponse(ctx, tmpl, item, res)
		}
	}
}

func (s *service) fetchItems(ctx context.Context, sessionCtxData *types.SessionContextData, req *http.Request) (items *types.ItemList, err error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger
	tracing.AttachRequestToSpan(span, req)

	if s.useFakeData {
		items = fakes.BuildFakeItemList()
	} else {
		filter := types.ExtractQueryFilter(req)
		items, err = s.dataStore.GetItems(ctx, sessionCtxData.ActiveAccountID, filter)
		if err != nil {
			return nil, observability.PrepareError(err, logger, span, "fetching item data")
		}
	}

	return items, nil
}

//go:embed templates/partials/generated/tables/items_table.gotpl
var itemsTableTemplate string

func (s *service) buildItemsTableView(includeBaseTemplate bool) func(http.ResponseWriter, *http.Request) {
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

		items, err := s.fetchItems(ctx, sessionCtxData, req)
		if err != nil {
			observability.AcknowledgeError(err, logger, span, "fetching items from datastore")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		tmplFuncMap := map[string]interface{}{
			"individualURL": func(x *types.Item) template.URL {
				// #nosec G203
				return template.URL(fmt.Sprintf("/dashboard_pages/items/%d", x.ID))
			},
			"pushURL": func(x *types.Item) template.URL {
				// #nosec G203
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
				page.IsServiceAdmin = sessionCtxData.Requester.ServicePermissions.IsServiceAdmin()
			}

			s.renderTemplateToResponse(ctx, tmpl, page, res)
		} else {
			tmpl := s.parseTemplate(ctx, "dashboard", itemsTableTemplate, tmplFuncMap)

			s.renderTemplateToResponse(ctx, tmpl, items, res)
		}
	}
}

// parseFormEncodedItemUpdateInput checks a request for an ItemUpdateInput.
func (s *service) parseFormEncodedItemUpdateInput(ctx context.Context, req *http.Request, sessionCtxData *types.SessionContextData) (updateInput *types.ItemUpdateInput) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	form, err := s.extractFormFromRequest(ctx, req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "parsing item creation input")
		return nil
	}

	updateInput = &types.ItemUpdateInput{
		Name:             form.Get(itemUpdateInputNameFormKey),
		Details:          form.Get(itemUpdateInputDetailsFormKey),
		BelongsToAccount: sessionCtxData.ActiveAccountID,
	}

	if err = updateInput.ValidateWithContext(ctx); err != nil {
		logger = logger.WithValue("input", updateInput)
		observability.AcknowledgeError(err, logger, span, "invalid item creation input")
		return nil
	}

	return updateInput
}

func (s *service) handleItemUpdateRequest(res http.ResponseWriter, req *http.Request) {
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

	if err = s.dataStore.UpdateItem(ctx, item, sessionCtxData.Requester.UserID, changes); err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching item from datastore")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	tmplFuncMap := map[string]interface{}{
		"componentTitle": func(x *types.Item) string {
			return fmt.Sprintf("Item #%d", x.ID)
		},
	}

	tmpl := s.parseTemplate(ctx, "", itemEditorTemplate, tmplFuncMap)

	s.renderTemplateToResponse(ctx, tmpl, item, res)
}

func (s *service) handleItemDeletionRequest(res http.ResponseWriter, req *http.Request) {
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

	itemID := s.routeParamManager.BuildRouteParamIDFetcher(logger, itemIDURLParamKey, "item")(req)
	if err = s.dataStore.ArchiveItem(ctx, itemID, sessionCtxData.ActiveAccountID, sessionCtxData.Requester.UserID); err != nil {
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
			// #nosec G203
			return template.URL(fmt.Sprintf("/dashboard_pages/items/%d", x.ID))
		},
		"pushURL": func(x *types.Item) template.URL {
			// #nosec G203
			return template.URL(fmt.Sprintf("/items/%d", x.ID))
		},
	}

	tmpl := s.parseTemplate(ctx, "dashboard", itemsTableTemplate, tmplFuncMap)

	s.renderTemplateToResponse(ctx, tmpl, items, res)
}
