package frontend

import (
	"context"
	// import embed for the side effect.
	_ "embed"
	"fmt"
	"html/template"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
)

const (
	userIDURLParamKey = "user"
)

func (s *Service) fetchUsers(ctx context.Context, req *http.Request) (users *types.UserList, err error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger

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

func (s *Service) buildUsersTableView(includeBaseTemplate, forSearch bool) func(http.ResponseWriter, *http.Request) {
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

		var users *types.UserList
		if forSearch {
			query := req.URL.Query().Get(types.SearchQueryKey)
			searchResults, err := s.dataStore.SearchForUsersByUsername(ctx, query)
			if err != nil {
				observability.AcknowledgeError(err, logger, span, "fetching users from datastore")
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

		tmplFuncMap := map[string]interface{}{
			"individualURL": func(x *types.User) template.URL {
				/* #nosec G203 */
				return template.URL(fmt.Sprintf("/dashboard_pages/admin/users/%d", x.ID))
			},
			"pushURL": func(x *types.User) template.URL {
				/* #nosec G203 */
				return template.URL(fmt.Sprintf("/admin/users/%d", x.ID))
			},
		}

		if includeBaseTemplate {
			tmpl := s.renderTemplateIntoBaseTemplate(usersTableTemplate, tmplFuncMap)

			page := &pageData{
				IsLoggedIn:  sessionCtxData != nil,
				Title:       "Users",
				ContentData: users,
			}
			if sessionCtxData != nil {
				page.IsServiceAdmin = sessionCtxData.Requester.ServiceAdminPermission.IsServiceAdmin()
			}

			if err = s.renderTemplateToResponse(tmpl, page, res); err != nil {
				observability.AcknowledgeError(err, logger, span, "rendering users dashboard tmpl")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			tmpl := s.parseTemplate("dashboard", usersTableTemplate, tmplFuncMap)

			if err = s.renderTemplateToResponse(tmpl, users, res); err != nil {
				observability.AcknowledgeError(err, logger, span, "rendering users table view")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}
}
