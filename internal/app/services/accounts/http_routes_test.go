package accounts

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAccountsService_ListHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleAccountList := fakes.BuildFakeAccountList()

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On(
			"GetAccounts",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleAccountList, nil)
		helper.service.accountDataManager = accountDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			mock.IsType(&types.AccountList{}),
		)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ListHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, accountDataManager, encoderDecoder)
	})

	T.Run("standard for admin", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleAccountList := fakes.BuildFakeAccountList()

		helper.req.URL.RawQuery = "admin=true"

		fart := helper.req.URL.String()
		_ = fart

		helper.exampleUser.ServiceAdminPermission = testutil.BuildMaxServiceAdminPerms()
		helper.service.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			sessionCtxData, err := types.SessionContextDataFromUser(
				helper.exampleUser,
				helper.exampleAccount.ID,
				map[uint64]*types.UserAccountMembershipInfo{
					helper.exampleAccount.ID: {
						AccountName: helper.exampleAccount.Name,
						Permissions: testutil.BuildMaxUserPerms(),
					},
				},
			)
			require.NoError(t, err)

			return sessionCtxData, nil
		}

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On(
			"GetAccountsForAdmin",
			testutil.ContextMatcher,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleAccountList, nil)
		helper.service.accountDataManager = accountDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			mock.IsType(&types.AccountList{}),
		)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ListHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, accountDataManager, encoderDecoder)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ListHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with now rows returned", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On(
			"GetAccounts",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.AccountList)(nil), sql.ErrNoRows)
		helper.service.accountDataManager = accountDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			mock.IsType(&types.AccountList{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ListHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, accountDataManager, encoderDecoder)
	})

	T.Run("with error fetching accounts from database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On(
			"GetAccounts",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.AccountList)(nil), errors.New("blah"))
		helper.service.accountDataManager = accountDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ListHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountDataManager, encoderDecoder)
	})
}

func TestAccountsService_CreateHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleInput := fakes.BuildFakeAccountCreationInputFromAccount(helper.exampleAccount)

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On(
			"CreateAccount",
			testutil.ContextMatcher,
			mock.IsType(&types.AccountCreationInput{}),
			helper.exampleUser.ID,
		).Return(helper.exampleAccount, nil)
		helper.service.accountDataManager = accountDataManager

		unitCounter := &mockmetrics.UnitCounter{}
		unitCounter.On("Increment", testutil.ContextMatcher).Return()
		helper.service.accountCounter = unitCounter

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeResponseWithStatus",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			mock.IsType(&types.Account{}),
			http.StatusCreated,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), createMiddlewareCtxKey, exampleInput))

		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusCreated, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountDataManager, unitCounter, encoderDecoder)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleInput := fakes.BuildFakeAccountCreationInputFromAccount(helper.exampleAccount)

		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), createMiddlewareCtxKey, exampleInput))
		helper.service.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			return nil, errors.New("blah")
		}

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("without input attached to request", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeInvalidInputResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with error creating account in database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleInput := fakes.BuildFakeAccountCreationInputFromAccount(helper.exampleAccount)

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On(
			"CreateAccount",
			testutil.ContextMatcher,
			mock.IsType(&types.AccountCreationInput{}),
			helper.exampleUser.ID,
		).Return((*types.Account)(nil), errors.New("blah"))
		helper.service.accountDataManager = accountDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), createMiddlewareCtxKey, exampleInput))

		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountDataManager, encoderDecoder)
	})
}

func TestAccountsService_ReadHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On(
			"GetAccount",
			testutil.ContextMatcher,
			helper.exampleAccount.ID,
			helper.exampleUser.ID,
		).Return(helper.exampleAccount, nil)
		helper.service.accountDataManager = accountDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			mock.IsType(&types.Account{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ReadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, accountDataManager, encoderDecoder)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.service.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			return nil, errors.New("blah")
		}

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ReadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with no such account in database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On(
			"GetAccount",
			testutil.ContextMatcher,
			helper.exampleAccount.ID,
			helper.exampleUser.ID,
		).Return((*types.Account)(nil), sql.ErrNoRows)
		helper.service.accountDataManager = accountDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ReadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountDataManager, encoderDecoder)
	})

	T.Run("with error reading from database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On(
			"GetAccount",
			testutil.ContextMatcher,
			helper.exampleAccount.ID, helper.exampleUser.ID,
		).Return((*types.Account)(nil), errors.New("blah"))
		helper.service.accountDataManager = accountDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ReadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountDataManager, encoderDecoder)
	})
}

func TestAccountsService_UpdateHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleInput := fakes.BuildFakeAccountUpdateInputFromAccount(helper.exampleAccount)

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On(
			"GetAccount",
			testutil.ContextMatcher,
			helper.exampleAccount.ID, helper.exampleUser.ID,
		).Return(helper.exampleAccount, nil)
		accountDataManager.On(
			"UpdateAccount",
			testutil.ContextMatcher,
			mock.IsType(&types.Account{}), helper.exampleUser.ID,
			mock.IsType([]*types.FieldChangeSummary{}),
		).Return(nil)
		helper.service.accountDataManager = accountDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			mock.IsType(&types.Account{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), updateMiddlewareCtxKey, exampleInput))

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, accountDataManager, helper.service.accountDataManager, encoderDecoder)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleInput := fakes.BuildFakeAccountUpdateInputFromAccount(helper.exampleAccount)

		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), updateMiddlewareCtxKey, exampleInput))
		helper.service.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			return nil, errors.New("blah")
		}

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)
	})

	T.Run("without update input attached to request", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeInvalidInputResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with no rows", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleInput := fakes.BuildFakeAccountUpdateInputFromAccount(helper.exampleAccount)

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On(
			"GetAccount",
			testutil.ContextMatcher,
			helper.exampleAccount.ID, helper.exampleUser.ID,
		).Return((*types.Account)(nil), sql.ErrNoRows)
		helper.service.accountDataManager = accountDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), updateMiddlewareCtxKey, exampleInput))

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountDataManager, encoderDecoder)
	})

	T.Run("with error querying for account", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleInput := fakes.BuildFakeAccountUpdateInputFromAccount(helper.exampleAccount)

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On(
			"GetAccount",
			testutil.ContextMatcher,
			helper.exampleAccount.ID, helper.exampleUser.ID,
		).Return((*types.Account)(nil), errors.New("blah"))
		helper.service.accountDataManager = accountDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), updateMiddlewareCtxKey, exampleInput))

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountDataManager, encoderDecoder)
	})

	T.Run("with error updating account", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.exampleAccount = fakes.BuildFakeAccount()
		helper.exampleAccount.BelongsToUser = helper.exampleUser.ID
		exampleInput := fakes.BuildFakeAccountUpdateInputFromAccount(helper.exampleAccount)

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On(
			"GetAccount",
			testutil.ContextMatcher,
			helper.exampleAccount.ID, helper.exampleUser.ID,
		).Return(helper.exampleAccount, nil)
		accountDataManager.On(
			"UpdateAccount",
			testutil.ContextMatcher,
			mock.IsType(&types.Account{}), helper.exampleUser.ID,
			mock.IsType([]*types.FieldChangeSummary{}),
		).Return(errors.New("blah"))
		helper.service.accountDataManager = accountDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), updateMiddlewareCtxKey, exampleInput))

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountDataManager, encoderDecoder)
	})
}

func TestAccountsService_ArchiveHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On(
			"ArchiveAccount",
			testutil.ContextMatcher,
			helper.exampleAccount.ID, helper.exampleUser.ID, helper.exampleUser.ID,
		).Return(nil)
		helper.service.accountDataManager = accountDataManager

		unitCounter := &mockmetrics.UnitCounter{}
		unitCounter.On("Decrement", testutil.ContextMatcher).Return()
		helper.service.accountCounter = unitCounter

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNoContent, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountDataManager, unitCounter)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			return nil, errors.New("blah")
		}

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with no such account in database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On(
			"ArchiveAccount",
			testutil.ContextMatcher,
			helper.exampleAccount.ID, helper.exampleUser.ID, helper.exampleUser.ID,
		).Return(sql.ErrNoRows)
		helper.service.accountDataManager = accountDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountDataManager, encoderDecoder)
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On(
			"ArchiveAccount",
			testutil.ContextMatcher,
			helper.exampleAccount.ID, helper.exampleUser.ID, helper.exampleUser.ID,
		).Return(errors.New("blah"))
		helper.service.accountDataManager = accountDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountDataManager, encoderDecoder)
	})
}

func TestAccountsService_AddUserHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleInput := fakes.BuildFakeAddUserToAccountInput()
		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), addUserToAccountMiddlewareCtxKey, exampleInput))

		accountMembershipDataManager := &mocktypes.AccountUserMembershipDataManager{}
		accountMembershipDataManager.On(
			"AddUserToAccount",
			testutil.ContextMatcher,
			exampleInput,
			helper.exampleUser.ID,
		).Return(nil)
		helper.service.accountMembershipDataManager = accountMembershipDataManager

		helper.service.AddUserHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusAccepted, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountMembershipDataManager)
	})

	T.Run("with missing input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeInvalidInputResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.AddUserHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		exampleInput := fakes.BuildFakeAddUserToAccountInput()
		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), addUserToAccountMiddlewareCtxKey, exampleInput))

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.AddUserHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleInput := fakes.BuildFakeAddUserToAccountInput()
		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), addUserToAccountMiddlewareCtxKey, exampleInput))

		accountMembershipDataManager := &mocktypes.AccountUserMembershipDataManager{}
		accountMembershipDataManager.On(
			"AddUserToAccount",
			testutil.ContextMatcher,
			exampleInput,
			helper.exampleUser.ID,
		).Return(errors.New("blah"))
		helper.service.accountMembershipDataManager = accountMembershipDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.AddUserHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountMembershipDataManager)
	})
}

func TestAccountsService_ModifyMemberPermissionsHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleInput := fakes.BuildFakeUserPermissionModificationInput()
		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), modifyMembershipMiddlewareCtxKey, exampleInput))

		accountMembershipDataManager := &mocktypes.AccountUserMembershipDataManager{}
		accountMembershipDataManager.On(
			"ModifyUserPermissions",
			testutil.ContextMatcher,
			helper.exampleAccount.ID,
			helper.exampleUser.ID,
			exampleInput,
		).Return(nil)
		helper.service.accountMembershipDataManager = accountMembershipDataManager

		helper.service.ModifyMemberPermissionsHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusAccepted, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountMembershipDataManager)
	})

	T.Run("with missing input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeInvalidInputResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ModifyMemberPermissionsHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		exampleInput := fakes.BuildFakeUserPermissionModificationInput()
		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), modifyMembershipMiddlewareCtxKey, exampleInput))

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ModifyMemberPermissionsHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleInput := fakes.BuildFakeUserPermissionModificationInput()
		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), modifyMembershipMiddlewareCtxKey, exampleInput))

		accountMembershipDataManager := &mocktypes.AccountUserMembershipDataManager{}
		accountMembershipDataManager.On(
			"ModifyUserPermissions",
			testutil.ContextMatcher,
			helper.exampleAccount.ID,
			helper.exampleUser.ID,
			exampleInput,
		).Return(errors.New("blah"))
		helper.service.accountMembershipDataManager = accountMembershipDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ModifyMemberPermissionsHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountMembershipDataManager, encoderDecoder)
	})
}

func TestAccountsService_TransferAccountOwnershipHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleInput := fakes.BuildFakeTransferAccountOwnershipInput()
		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), transferAccountMiddlewareCtxKey, exampleInput))

		accountMembershipDataManager := &mocktypes.AccountUserMembershipDataManager{}
		accountMembershipDataManager.On(
			"TransferAccountOwnership",
			testutil.ContextMatcher,
			helper.exampleAccount.ID,
			helper.exampleUser.ID,
			exampleInput,
		).Return(nil)
		helper.service.accountMembershipDataManager = accountMembershipDataManager

		helper.service.TransferAccountOwnershipHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusAccepted, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountMembershipDataManager)
	})

	T.Run("without input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeInvalidInputResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.TransferAccountOwnershipHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		exampleInput := fakes.BuildFakeTransferAccountOwnershipInput()
		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), transferAccountMiddlewareCtxKey, exampleInput))

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.TransferAccountOwnershipHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleInput := fakes.BuildFakeTransferAccountOwnershipInput()
		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), transferAccountMiddlewareCtxKey, exampleInput))

		accountMembershipDataManager := &mocktypes.AccountUserMembershipDataManager{}
		accountMembershipDataManager.On(
			"TransferAccountOwnership",
			testutil.ContextMatcher,
			helper.exampleAccount.ID,
			helper.exampleUser.ID,
			exampleInput,
		).Return(errors.New("blah"))
		helper.service.accountMembershipDataManager = accountMembershipDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.TransferAccountOwnershipHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountMembershipDataManager, encoderDecoder)
	})
}

func TestAccountsService_RemoveUserHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleReason := t.Name()

		accountMembershipDataManager := &mocktypes.AccountUserMembershipDataManager{}
		accountMembershipDataManager.On(
			"RemoveUserFromAccount",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
			helper.exampleAccount.ID,
			helper.exampleUser.ID,
			exampleReason,
		).Return(nil)
		helper.service.accountMembershipDataManager = accountMembershipDataManager

		helper.req.URL.RawQuery = fmt.Sprintf("reason=%s", url.QueryEscape(exampleReason))

		helper.service.RemoveUserHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusAccepted, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountMembershipDataManager)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.RemoveUserHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleReason := t.Name()
		helper.req.URL.RawQuery = fmt.Sprintf("reason=%s", url.QueryEscape(exampleReason))

		accountMembershipDataManager := &mocktypes.AccountUserMembershipDataManager{}
		accountMembershipDataManager.On(
			"RemoveUserFromAccount",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
			helper.exampleAccount.ID,
			helper.exampleUser.ID,
			exampleReason,
		).Return(errors.New("blah"))
		helper.service.accountMembershipDataManager = accountMembershipDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.RemoveUserHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountMembershipDataManager, encoderDecoder)
	})
}

func TestAccountsService_MarkAsDefaultHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		accountMembershipDataManager := &mocktypes.AccountUserMembershipDataManager{}
		accountMembershipDataManager.On(
			"MarkAccountAsUserDefault",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
			helper.exampleAccount.ID,
			helper.exampleUser.ID,
		).Return(nil)
		helper.service.accountMembershipDataManager = accountMembershipDataManager

		helper.service.MarkAsDefaultHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusAccepted, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountMembershipDataManager)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.MarkAsDefaultHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		accountMembershipDataManager := &mocktypes.AccountUserMembershipDataManager{}
		accountMembershipDataManager.On(
			"MarkAccountAsUserDefault",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
			helper.exampleAccount.ID,
			helper.exampleUser.ID,
		).Return(errors.New("blah"))
		helper.service.accountMembershipDataManager = accountMembershipDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.MarkAsDefaultHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountMembershipDataManager, encoderDecoder)
	})
}

func TestAccountsService_AuditEntryHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleAuditLogEntries := fakes.BuildFakeAuditLogEntryList().Entries

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On(
			"GetAuditLogEntriesForAccount",
			testutil.ContextMatcher,
			helper.exampleAccount.ID,
		).Return(exampleAuditLogEntries, nil)
		helper.service.accountDataManager = accountDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			mock.IsType([]*types.AuditLogEntry{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.AuditEntryHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountDataManager, encoderDecoder)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.AuditEntryHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On(
			"GetAuditLogEntriesForAccount",
			testutil.ContextMatcher,
			helper.exampleAccount.ID,
		).Return([]*types.AuditLogEntry(nil), sql.ErrNoRows)
		helper.service.accountDataManager = accountDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.AuditEntryHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountDataManager, encoderDecoder)
	})

	T.Run("with error reading from database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		accountDataManager := &mocktypes.AccountDataManager{}
		accountDataManager.On(
			"GetAuditLogEntriesForAccount",
			testutil.ContextMatcher,
			helper.exampleAccount.ID,
		).Return([]*types.AuditLogEntry(nil), errors.New("blah"))
		helper.service.accountDataManager = accountDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.AuditEntryHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountDataManager, encoderDecoder)
	})
}
