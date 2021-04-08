package auth

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"
)

func TestAuthService_DecodeCookieFromRequest(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.ctx, helper.req = attachCookieToRequestForTest(t, helper.service, helper.req, helper.exampleUser)

		_, userID, err := helper.service.getUserIDFromCookie(helper.ctx, helper.req)
		assert.NoError(t, err)
		assert.Equal(t, helper.exampleUser.ID, userID)
	})

	T.Run("with invalid cookie", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		// begin building bad cookie.
		// NOTE: any code here is duplicated from service.buildAuthCookie
		// any changes made there might need to be reflected here.
		c := &http.Cookie{
			Name:     helper.service.config.Cookies.Name,
			Value:    "blah blah blah this is not a real cookie",
			Path:     "/",
			HttpOnly: true,
		}
		// end building bad cookie.
		helper.req.AddCookie(c)

		_, userID, err := helper.service.getUserIDFromCookie(helper.req.Context(), helper.req)
		assert.Error(t, err)
		assert.Zero(t, userID)
	})

	T.Run("without cookie", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		_, userID, err := helper.service.getUserIDFromCookie(helper.req.Context(), helper.req)
		assert.Error(t, err)
		assert.Equal(t, err, http.ErrNoCookie)
		assert.Zero(t, userID)
	})
}

func TestAuthService_determineUserFromRequestCookie(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.ctx, helper.req = attachCookieToRequestForTest(t, helper.service, helper.req, helper.exampleUser)

		udb := &mocktypes.UserDataManager{}
		udb.On(
			"GetUser",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = udb

		actualUser, err := helper.service.determineUserFromRequestCookie(helper.ctx, helper.req)
		assert.Equal(t, helper.exampleUser, actualUser)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, udb)
	})

	T.Run("without cookie", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		actualUser, err := helper.service.determineUserFromRequestCookie(helper.req.Context(), helper.req)
		assert.Nil(t, actualUser)
		assert.Error(t, err)
	})

	T.Run("with error retrieving user from datastore", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		helper.ctx, helper.req = attachCookieToRequestForTest(t, helper.service, helper.req, helper.exampleUser)

		expectedError := errors.New("blah")
		udb := &mocktypes.UserDataManager{}
		udb.On(
			"GetUser",
			mock.MatchedBy(testutil.ContextMatcher),
			helper.exampleUser.ID,
		).Return((*types.User)(nil), expectedError)
		helper.service.userDataManager = udb

		actualUser, err := helper.service.determineUserFromRequestCookie(helper.req.Context(), helper.req)
		assert.Nil(t, actualUser)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, udb)
	})
}
