package accountsubscriptionplans

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"testing"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"

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
		accountSubscriptionPlanDataManager.On("GetAccountSubscriptionPlans", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.QueryFilter{})).Return(exampleAccountSubscriptionPlanList, nil)
		helper.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.AccountSubscriptionPlanList{}))
		helper.service.encoderDecoder = ed

		helper.service.ListHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d ins.service.atus response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, ed)
	})

	T.Run("with no rows returned", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		accountSubscriptionPlanDataManager.On("GetAccountSubscriptionPlans", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.QueryFilter{})).Return((*types.AccountSubscriptionPlanList)(nil), sql.ErrNoRows)
		helper.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.AccountSubscriptionPlanList{}))
		helper.service.encoderDecoder = ed

		helper.service.ListHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected response status to be %d, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, ed)
	})

	T.Run("with error fetching account subscription plans from database", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		accountSubscriptionPlanDataManager.On("GetAccountSubscriptionPlans", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.QueryFilter{})).Return((*types.AccountSubscriptionPlanList)(nil), errors.New("blah"))
		helper.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.ListHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, ed)
	})
}

func TestAccountSubscriptionPlansService_CreateHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		exampleInput := fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(helper.exampleAccountSubscriptionPlan)

		accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		accountSubscriptionPlanDataManager.On("CreateAccountSubscriptionPlan", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.AccountSubscriptionPlanCreationInput{})).Return(helper.exampleAccountSubscriptionPlan, nil)
		helper.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

		mc := &mockmetrics.UnitCounter{}
		mc.On("Increment", mock.MatchedBy(testutil.ContextMatcher))
		helper.service.planCounter = mc

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeResponseWithStatus", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.AccountSubscriptionPlan{}), http.StatusCreated)
		helper.service.encoderDecoder = ed

		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), createMiddlewareCtxKey, exampleInput))

		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusCreated, helper.res.Code)

		mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, mc, ed)
	})

	T.Run("without input attached to request", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeInvalidInputResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)

		mock.AssertExpectationsForObjects(t, ed)
	})

	T.Run("with error creating account subscription plan in database", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		exampleInput := fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(helper.exampleAccountSubscriptionPlan)

		accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		accountSubscriptionPlanDataManager.On("CreateAccountSubscriptionPlan", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.AccountSubscriptionPlanCreationInput{})).Return((*types.AccountSubscriptionPlan)(nil), errors.New("blah"))
		helper.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), createMiddlewareCtxKey, exampleInput))

		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, ed)
	})
}

func TestAccountSubscriptionPlansService_ReadHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		accountSubscriptionPlanDataManager.On("GetAccountSubscriptionPlan", mock.MatchedBy(testutil.ContextMatcher), helper.exampleAccountSubscriptionPlan.ID).Return(helper.exampleAccountSubscriptionPlan, nil)
		helper.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.AccountSubscriptionPlan{}))
		helper.service.encoderDecoder = ed

		helper.service.ReadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d ins.service.atus response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, ed)
	})

	T.Run("with no such account subscription plan in the database", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		accountSubscriptionPlanDataManager.On("GetAccountSubscriptionPlan", mock.MatchedBy(testutil.ContextMatcher), helper.exampleAccountSubscriptionPlan.ID).Return((*types.AccountSubscriptionPlan)(nil), sql.ErrNoRows)
		helper.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeNotFoundResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.ReadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)

		mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, ed)
	})

	T.Run("with error fetching account subscription plan from database", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		accountSubscriptionPlanDataManager.On("GetAccountSubscriptionPlan", mock.MatchedBy(testutil.ContextMatcher), helper.exampleAccountSubscriptionPlan.ID).Return((*types.AccountSubscriptionPlan)(nil), errors.New("blah"))
		helper.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.ReadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, ed)
	})
}

func TestAccountSubscriptionPlansService_UpdateHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		exampleInput := fakes.BuildFakePlanUpdateInputFromPlan(helper.exampleAccountSubscriptionPlan)

		accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		accountSubscriptionPlanDataManager.On("GetAccountSubscriptionPlan", mock.MatchedBy(testutil.ContextMatcher), helper.exampleAccountSubscriptionPlan.ID).Return(helper.exampleAccountSubscriptionPlan, nil)
		accountSubscriptionPlanDataManager.On("UpdateAccountSubscriptionPlan", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.AccountSubscriptionPlan{})).Return(nil)
		helper.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.AccountSubscriptionPlan{}))
		helper.service.encoderDecoder = ed

		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), updateMiddlewareCtxKey, exampleInput))

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d ins.service.atus response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, ed)
	})

	T.Run("without input attached to request", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeInvalidInputResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)

		mock.AssertExpectationsForObjects(t, ed)
	})

	T.Run("with no results in database", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		exampleInput := fakes.BuildFakePlanUpdateInputFromPlan(helper.exampleAccountSubscriptionPlan)

		accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		accountSubscriptionPlanDataManager.On("GetAccountSubscriptionPlan", mock.MatchedBy(testutil.ContextMatcher), helper.exampleAccountSubscriptionPlan.ID).Return((*types.AccountSubscriptionPlan)(nil), sql.ErrNoRows)
		helper.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeNotFoundResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), updateMiddlewareCtxKey, exampleInput))

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)

		mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, ed)
	})

	T.Run("with error fetching from database", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		exampleInput := fakes.BuildFakePlanUpdateInputFromPlan(helper.exampleAccountSubscriptionPlan)

		accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		accountSubscriptionPlanDataManager.On("GetAccountSubscriptionPlan", mock.MatchedBy(testutil.ContextMatcher), helper.exampleAccountSubscriptionPlan.ID).Return((*types.AccountSubscriptionPlan)(nil), errors.New("blah"))
		helper.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), updateMiddlewareCtxKey, exampleInput))

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, ed)
	})

	T.Run("with error performing update", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		exampleInput := fakes.BuildFakePlanUpdateInputFromPlan(helper.exampleAccountSubscriptionPlan)

		accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		accountSubscriptionPlanDataManager.On("GetAccountSubscriptionPlan", mock.MatchedBy(testutil.ContextMatcher), helper.exampleAccountSubscriptionPlan.ID).Return(helper.exampleAccountSubscriptionPlan, nil)
		accountSubscriptionPlanDataManager.On("UpdateAccountSubscriptionPlan", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.AccountSubscriptionPlan{})).Return(errors.New("blah"))
		helper.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), updateMiddlewareCtxKey, exampleInput))

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, ed)
	})
}

func TestAccountSubscriptionPlansService_ArchiveHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		accountSubscriptionPlanDataManager.On("ArchiveAccountSubscriptionPlan", mock.MatchedBy(testutil.ContextMatcher), helper.exampleAccountSubscriptionPlan.ID, helper.exampleUser.ID).Return(nil)
		helper.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

		mc := &mockmetrics.UnitCounter{}
		mc.On("Decrement", mock.MatchedBy(testutil.ContextMatcher)).Return()
		helper.service.planCounter = mc

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNoContent, helper.res.Code)

		mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, mc)
	})

	T.Run("with no such account subscription plan in the database", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		accountSubscriptionPlanDataManager.On("ArchiveAccountSubscriptionPlan", mock.MatchedBy(testutil.ContextMatcher), helper.exampleAccountSubscriptionPlan.ID, helper.exampleUser.ID).Return(sql.ErrNoRows)
		helper.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeNotFoundResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)

		mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, ed)
	})

	T.Run("with error marking as archived in database", func(t *testing.T) {
		t.Parallel()
		helper := buildTestHelper(t)

		accountSubscriptionPlanDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		accountSubscriptionPlanDataManager.On("ArchiveAccountSubscriptionPlan", mock.MatchedBy(testutil.ContextMatcher), helper.exampleAccountSubscriptionPlan.ID, helper.exampleUser.ID).Return(errors.New("blah"))
		helper.service.accountSubscriptionPlanDataManager = accountSubscriptionPlanDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		helper.service.encoderDecoder = ed

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, accountSubscriptionPlanDataManager, ed)
	})
}
