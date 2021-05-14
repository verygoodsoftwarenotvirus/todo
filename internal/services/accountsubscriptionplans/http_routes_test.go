package accountsubscriptionplans

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"testing"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/mock"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/metrics/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/mock"
	testutil "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAccountSubscriptionPlansService_ListHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleAccountSubscriptionPlanList := fakes.BuildFakeAccountSubscriptionPlanList()

		accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		accountSubscriptionPlanDataManager.On(
			"GetAccountSubscriptionPlans",
			testutil.ContextMatcher,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleAccountSubscriptionPlanList, nil)
		helper.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			mock.IsType(&types.AccountSubscriptionPlanList{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ListHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, encoderDecoder)
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
		)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ListHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with no rows returned", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		accountSubscriptionPlanDataManager.On(
			"GetAccountSubscriptionPlans",
			testutil.ContextMatcher,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.AccountSubscriptionPlanList)(nil), sql.ErrNoRows)
		helper.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			mock.IsType(&types.AccountSubscriptionPlanList{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ListHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected response status to be %d, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, encoderDecoder)
	})

	T.Run("with error fetching account subscription plans from database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		accountSubscriptionPlanDataManager.On(
			"GetAccountSubscriptionPlans",
			testutil.ContextMatcher,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.AccountSubscriptionPlanList)(nil), errors.New("blah"))
		helper.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ListHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, encoderDecoder)
	})
}

func TestAccountSubscriptionPlansService_CreateHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleInput := fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(helper.exampleAccountSubscriptionPlan)
		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), createMiddlewareCtxKey, exampleInput))

		accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		accountSubscriptionPlanDataManager.On(
			"CreateAccountSubscriptionPlan",
			testutil.ContextMatcher,
			mock.IsType(&types.AccountSubscriptionPlanCreationInput{}),
		).Return(helper.exampleAccountSubscriptionPlan, nil)
		helper.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

		unitCounter := &mockmetrics.UnitCounter{}
		unitCounter.On("Increment", testutil.ContextMatcher).Return()
		helper.service.planCounter = unitCounter

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeResponseWithStatus",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			mock.IsType(&types.AccountSubscriptionPlan{}),
			http.StatusCreated,
		)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusCreated, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, unitCounter, encoderDecoder)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleInput := fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(helper.exampleAccountSubscriptionPlan)
		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), createMiddlewareCtxKey, exampleInput))

		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		)
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

	T.Run("with error creating account subscription plan in database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleInput := fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(helper.exampleAccountSubscriptionPlan)

		accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		accountSubscriptionPlanDataManager.On(
			"CreateAccountSubscriptionPlan",
			testutil.ContextMatcher,
			mock.IsType(&types.AccountSubscriptionPlanCreationInput{}),
		).Return((*types.AccountSubscriptionPlan)(nil), errors.New("blah"))
		helper.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher)
		helper.service.encoderDecoder = encoderDecoder

		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), createMiddlewareCtxKey, exampleInput))

		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, encoderDecoder)
	})
}

func TestAccountSubscriptionPlansService_ReadHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		accountSubscriptionPlanDataManager.On(
			"GetAccountSubscriptionPlan",
			testutil.ContextMatcher,
			helper.exampleAccountSubscriptionPlan.ID,
		).Return(helper.exampleAccountSubscriptionPlan, nil)
		helper.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			mock.IsType(&types.AccountSubscriptionPlan{}))
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ReadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, encoderDecoder)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleInput := fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(helper.exampleAccountSubscriptionPlan)
		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), createMiddlewareCtxKey, exampleInput))

		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ReadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with no such account subscription plan in the database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		accountSubscriptionPlanDataManager.On(
			"GetAccountSubscriptionPlan",
			testutil.ContextMatcher,
			helper.exampleAccountSubscriptionPlan.ID,
		).Return((*types.AccountSubscriptionPlan)(nil), sql.ErrNoRows)
		helper.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ReadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, encoderDecoder)
	})

	T.Run("with error fetching account subscription plan from database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		accountSubscriptionPlanDataManager.On(
			"GetAccountSubscriptionPlan",
			testutil.ContextMatcher,
			helper.exampleAccountSubscriptionPlan.ID,
		).Return((*types.AccountSubscriptionPlan)(nil), errors.New("blah"))
		helper.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ReadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, encoderDecoder)
	})
}

func TestAccountSubscriptionPlansService_UpdateHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleInput := fakes.BuildFakePlanUpdateInputFromPlan(helper.exampleAccountSubscriptionPlan)

		accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		accountSubscriptionPlanDataManager.On("GetAccountSubscriptionPlan", testutil.ContextMatcher, helper.exampleAccountSubscriptionPlan.ID).Return(helper.exampleAccountSubscriptionPlan, nil)
		accountSubscriptionPlanDataManager.On("UpdateAccountSubscriptionPlan", testutil.ContextMatcher, mock.IsType(&types.AccountSubscriptionPlan{})).Return(nil)
		helper.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On("RespondWithData", testutil.ContextMatcher, testutil.ResponseWriterMatcher, mock.IsType(&types.AccountSubscriptionPlan{}))
		helper.service.encoderDecoder = encoderDecoder

		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), updateMiddlewareCtxKey, exampleInput))

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, encoderDecoder)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		exampleInput := fakes.BuildFakePlanUpdateInputFromPlan(helper.exampleAccountSubscriptionPlan)
		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), updateMiddlewareCtxKey, exampleInput))

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("without input attached to request", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeInvalidInputResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with no results in database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleInput := fakes.BuildFakePlanUpdateInputFromPlan(helper.exampleAccountSubscriptionPlan)

		accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		accountSubscriptionPlanDataManager.On(
			"GetAccountSubscriptionPlan",
			testutil.ContextMatcher,
			helper.exampleAccountSubscriptionPlan.ID,
		).Return((*types.AccountSubscriptionPlan)(nil), sql.ErrNoRows)
		helper.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher)
		helper.service.encoderDecoder = encoderDecoder

		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), updateMiddlewareCtxKey, exampleInput))

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, encoderDecoder)
	})

	T.Run("with error fetching from database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleInput := fakes.BuildFakePlanUpdateInputFromPlan(helper.exampleAccountSubscriptionPlan)

		accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		accountSubscriptionPlanDataManager.On(
			"GetAccountSubscriptionPlan",
			testutil.ContextMatcher,
			helper.exampleAccountSubscriptionPlan.ID,
		).Return((*types.AccountSubscriptionPlan)(nil), errors.New("blah"))
		helper.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher)
		helper.service.encoderDecoder = encoderDecoder

		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), updateMiddlewareCtxKey, exampleInput))

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, encoderDecoder)
	})

	T.Run("with error performing update", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleInput := fakes.BuildFakePlanUpdateInputFromPlan(helper.exampleAccountSubscriptionPlan)

		accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		accountSubscriptionPlanDataManager.On(
			"GetAccountSubscriptionPlan",
			testutil.ContextMatcher,
			helper.exampleAccountSubscriptionPlan.ID,
		).Return(helper.exampleAccountSubscriptionPlan, nil)
		accountSubscriptionPlanDataManager.On(
			"UpdateAccountSubscriptionPlan",
			testutil.ContextMatcher,
			mock.IsType(&types.AccountSubscriptionPlan{}),
		).Return(errors.New("blah"))
		helper.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher)
		helper.service.encoderDecoder = encoderDecoder

		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), updateMiddlewareCtxKey, exampleInput))

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, encoderDecoder)
	})
}

func TestAccountSubscriptionPlansService_ArchiveHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		accountSubscriptionPlanDataManager.On(
			"ArchiveAccountSubscriptionPlan",
			testutil.ContextMatcher,
			helper.exampleAccountSubscriptionPlan.ID, helper.exampleUser.ID,
		).Return(nil)
		helper.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

		unitCounter := &mockmetrics.UnitCounter{}
		unitCounter.On("Decrement", testutil.ContextMatcher).Return()
		helper.service.planCounter = unitCounter

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNoContent, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, unitCounter)
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
		)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with no such account subscription plan in the database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		accountSubscriptionPlanDataManager.On(
			"ArchiveAccountSubscriptionPlan",
			testutil.ContextMatcher,
			helper.exampleAccountSubscriptionPlan.ID, helper.exampleUser.ID,
		).Return(sql.ErrNoRows)
		helper.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, encoderDecoder)
	})

	T.Run("with error marking as archived in database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		accountSubscriptionPlanDataManager.On(
			"ArchiveAccountSubscriptionPlan",
			testutil.ContextMatcher,
			helper.exampleAccountSubscriptionPlan.ID, helper.exampleUser.ID,
		).Return(errors.New("blah"))
		helper.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, encoderDecoder)
	})
}

func TestAccountSubscriptionPlansService_AuditEntryHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleAuditLogEntries := fakes.BuildFakeAuditLogEntryList().Entries

		accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		accountSubscriptionPlanDataManager.On(
			"GetAuditLogEntriesForAccountSubscriptionPlan",
			testutil.ContextMatcher,
			helper.exampleAccountSubscriptionPlan.ID,
		).Return(exampleAuditLogEntries, nil)
		helper.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

		helper.service.AuditEntryHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager)
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

		accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		accountSubscriptionPlanDataManager.On(
			"GetAuditLogEntriesForAccountSubscriptionPlan",
			testutil.ContextMatcher,
			helper.exampleAccountSubscriptionPlan.ID,
		).Return([]*types.AuditLogEntry(nil), sql.ErrNoRows)
		helper.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

		helper.service.AuditEntryHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager)
	})

	T.Run("with error reading from database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		accountSubscriptionPlanDataManager.On(
			"GetAuditLogEntriesForAccountSubscriptionPlan",
			testutil.ContextMatcher,
			helper.exampleAccountSubscriptionPlan.ID,
		).Return([]*types.AuditLogEntry(nil), errors.New("blah"))
		helper.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

		helper.service.AuditEntryHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager)
	})
}
