package frontend

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"

	// import embed for the side effect.
	_ "embed"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
)

func (s *service) fetchUsers(ctx context.Context, req *http.Request) (users *types.UserList, err error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger
	tracing.AttachRequestToSpan(span, req)

	if s.useFakeData {
		users = fakes.BuildFakeUserList()
	} else {
		filter := types.ExtractQueryFilter(req)
		users, err = s.dataStore.GetUsers(ctx, filter)

		if err != nil {
			return nil, observability.PrepareError(err, logger, span, "fetching user data")
		}
	}

	return users, nil
}

//go:embed templates/partials/generated/tables/users_table.gotpl
var usersTableTemplate string

func (s *service) buildUsersTableView(includeBaseTemplate, forSearch bool) func(http.ResponseWriter, *http.Request) {
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

		var users *types.UserList
		if forSearch {
			query := req.URL.Query().Get(types.SearchQueryKey)
			searchResults, searchResultsErr := s.dataStore.SearchForUsersByUsername(ctx, query)
			if searchResultsErr != nil {
				observability.AcknowledgeError(searchResultsErr, logger, span, "fetching users from datastore")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}

			users = &types.UserList{Users: searchResults}
		} else {
			users, err = s.fetchUsers(ctx, req)
			if err != nil {
				observability.AcknowledgeError(err, logger, span, "fetching users from datastore")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		if includeBaseTemplate {
			tmpl := s.renderTemplateIntoBaseTemplate(usersTableTemplate, nil)

			page := &pageData{
				IsLoggedIn:  sessionCtxData != nil,
				Title:       "Users",
				ContentData: users,
			}
			if sessionCtxData != nil {
				page.IsServiceAdmin = sessionCtxData.Requester.ServiceRole.IsServiceAdmin()
			}

			s.renderTemplateToResponse(ctx, tmpl, page, res)
		} else {
			tmpl := s.parseTemplate(ctx, "dashboard", usersTableTemplate, nil)

			s.renderTemplateToResponse(ctx, tmpl, users, res)
		}
	}
}
