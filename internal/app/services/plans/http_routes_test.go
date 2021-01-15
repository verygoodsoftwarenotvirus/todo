package plans

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPlansService_ListHandler(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()
	sessionInfoFetcher := func(_ *http.Request) (*types.SessionInfo, error) {
		return exampleUser.ToSessionInfo(), nil
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		examplePlanList := fakes.BuildFakePlanList()

		planDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		planDataManager.On("GetAccountSubscriptionPlans", mock.Anything, mock.AnythingOfType("*types.QueryFilter")).Return(examplePlanList, nil)
		s.planDataManager = planDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything, mock.AnythingOfType("*types.AccountSubscriptionPlanList"))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ListHandler(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, planDataManager, ed)
	})

	T.Run("with no rows returned", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		planDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		planDataManager.On("GetAccountSubscriptionPlans", mock.Anything, mock.AnythingOfType("*types.QueryFilter")).Return((*types.AccountSubscriptionPlanList)(nil), sql.ErrNoRows)
		s.planDataManager = planDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything, mock.AnythingOfType("*types.AccountSubscriptionPlanList"))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ListHandler(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, planDataManager, ed)
	})

	T.Run("with error fetching plans from database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		planDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		planDataManager.On("GetAccountSubscriptionPlans", mock.Anything, mock.AnythingOfType("*types.QueryFilter")).Return((*types.AccountSubscriptionPlanList)(nil), errors.New("blah"))
		s.planDataManager = planDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything, mock.Anything)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ListHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, planDataManager, ed)
	})
}

func TestPlansService_CreateHandler(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()
	sessionInfoFetcher := func(_ *http.Request) (*types.SessionInfo, error) {
		return exampleUser.ToSessionInfo(), nil
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		examplePlan := fakes.BuildFakePlan()
		exampleInput := fakes.BuildFakePlanCreationInputFromPlan(examplePlan)

		planDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		planDataManager.On("CreateAccountSubscriptionPlan", mock.Anything, mock.AnythingOfType("*types.AccountSubscriptionPlanCreationInput")).Return(examplePlan, nil)
		s.planDataManager = planDataManager

		mc := &mockmetrics.UnitCounter{}
		mc.On("Increment", mock.Anything)
		s.planCounter = mc

		auditLog := &mocktypes.AuditLogEntryDataManager{}
		auditLog.On("LogAccountSubscriptionPlanCreationEvent", mock.Anything, examplePlan)
		s.auditLog = auditLog

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponseWithStatus", mock.Anything, mock.Anything, mock.AnythingOfType("*types.AccountSubscriptionPlan"), http.StatusCreated)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		req = req.WithContext(context.WithValue(req.Context(), createMiddlewareCtxKey, exampleInput))

		s.CreateHandler(res, req)

		assert.Equal(t, http.StatusCreated, res.Code)

		mock.AssertExpectationsForObjects(t, planDataManager, mc, ed)
	})

	T.Run("without input attached", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeInvalidInputResponse", mock.Anything, mock.Anything)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.CreateHandler(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)

		mock.AssertExpectationsForObjects(t, ed)
	})

	T.Run("with error creating plan", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		examplePlan := fakes.BuildFakePlan()
		exampleInput := fakes.BuildFakePlanCreationInputFromPlan(examplePlan)

		planDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		planDataManager.On("CreateAccountSubscriptionPlan", mock.Anything, mock.AnythingOfType("*types.AccountSubscriptionPlanCreationInput")).Return((*types.AccountSubscriptionPlan)(nil), errors.New("blah"))
		s.planDataManager = planDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything, mock.Anything)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		req = req.WithContext(context.WithValue(req.Context(), createMiddlewareCtxKey, exampleInput))

		s.CreateHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, planDataManager, ed)
	})
}

func TestPlansService_ReadHandler(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()
	sessionInfoFetcher := func(_ *http.Request) (*types.SessionInfo, error) {
		return exampleUser.ToSessionInfo(), nil
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		examplePlan := fakes.BuildFakePlan()
		s.planIDFetcher = func(req *http.Request) uint64 {
			return examplePlan.ID
		}

		planDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		planDataManager.On("GetAccountSubscriptionPlan", mock.Anything, examplePlan.ID).Return(examplePlan, nil)
		s.planDataManager = planDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything, mock.AnythingOfType("*types.AccountSubscriptionPlan"))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ReadHandler(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, planDataManager, ed)
	})

	T.Run("with no such plan in database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		examplePlan := fakes.BuildFakePlan()
		s.planIDFetcher = func(req *http.Request) uint64 {
			return examplePlan.ID
		}

		planDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		planDataManager.On("GetAccountSubscriptionPlan", mock.Anything, examplePlan.ID).Return((*types.AccountSubscriptionPlan)(nil), sql.ErrNoRows)
		s.planDataManager = planDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeNotFoundResponse", mock.Anything, mock.Anything)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ReadHandler(res, req)

		assert.Equal(t, http.StatusNotFound, res.Code)

		mock.AssertExpectationsForObjects(t, planDataManager, ed)
	})

	T.Run("with error fetching plan from database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		examplePlan := fakes.BuildFakePlan()
		s.planIDFetcher = func(req *http.Request) uint64 {
			return examplePlan.ID
		}

		planDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		planDataManager.On("GetAccountSubscriptionPlan", mock.Anything, examplePlan.ID).Return((*types.AccountSubscriptionPlan)(nil), errors.New("blah"))
		s.planDataManager = planDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything, mock.Anything)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ReadHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, planDataManager, ed)
	})
}

func TestPlansService_UpdateHandler(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()
	sessionInfoFetcher := func(_ *http.Request) (*types.SessionInfo, error) {
		return exampleUser.ToSessionInfo(), nil
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		examplePlan := fakes.BuildFakePlan()
		exampleInput := fakes.BuildFakePlanUpdateInputFromPlan(examplePlan)

		s.planIDFetcher = func(req *http.Request) uint64 {
			return examplePlan.ID
		}

		planDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		planDataManager.On("GetAccountSubscriptionPlan", mock.Anything, examplePlan.ID).Return(examplePlan, nil)
		planDataManager.On("UpdateAccountSubscriptionPlan", mock.Anything, mock.AnythingOfType("*types.AccountSubscriptionPlan")).Return(nil)
		s.planDataManager = planDataManager

		auditLog := &mocktypes.AuditLogEntryDataManager{}
		auditLog.On("AccountSubscriptionLogPlanUpdateEvent", mock.Anything, exampleUser.ID, examplePlan.ID, mock.AnythingOfType("[]types.FieldChangeSummary"))
		s.auditLog = auditLog

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything, mock.AnythingOfType("*types.AccountSubscriptionPlan"))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		req = req.WithContext(context.WithValue(req.Context(), updateMiddlewareCtxKey, exampleInput))

		s.UpdateHandler(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, planDataManager, ed)
	})

	T.Run("without update input", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeInvalidInputResponse", mock.Anything, mock.Anything)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.UpdateHandler(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)

		mock.AssertExpectationsForObjects(t, ed)
	})

	T.Run("with no rows fetching plan", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		examplePlan := fakes.BuildFakePlan()
		exampleInput := fakes.BuildFakePlanUpdateInputFromPlan(examplePlan)

		s.planIDFetcher = func(req *http.Request) uint64 {
			return examplePlan.ID
		}

		planDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		planDataManager.On("GetAccountSubscriptionPlan", mock.Anything, examplePlan.ID).Return((*types.AccountSubscriptionPlan)(nil), sql.ErrNoRows)
		s.planDataManager = planDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeNotFoundResponse", mock.Anything, mock.Anything)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		req = req.WithContext(context.WithValue(req.Context(), updateMiddlewareCtxKey, exampleInput))

		s.UpdateHandler(res, req)

		assert.Equal(t, http.StatusNotFound, res.Code)

		mock.AssertExpectationsForObjects(t, planDataManager, ed)
	})

	T.Run("with error fetching plan", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		examplePlan := fakes.BuildFakePlan()
		exampleInput := fakes.BuildFakePlanUpdateInputFromPlan(examplePlan)

		s.planIDFetcher = func(req *http.Request) uint64 {
			return examplePlan.ID
		}

		planDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		planDataManager.On("GetAccountSubscriptionPlan", mock.Anything, examplePlan.ID).Return((*types.AccountSubscriptionPlan)(nil), errors.New("blah"))
		s.planDataManager = planDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything, mock.Anything)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		req = req.WithContext(context.WithValue(req.Context(), updateMiddlewareCtxKey, exampleInput))

		s.UpdateHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, planDataManager, ed)
	})

	T.Run("with error updating plan", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		examplePlan := fakes.BuildFakePlan()
		exampleInput := fakes.BuildFakePlanUpdateInputFromPlan(examplePlan)

		s.planIDFetcher = func(req *http.Request) uint64 {
			return examplePlan.ID
		}

		planDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		planDataManager.On("GetAccountSubscriptionPlan", mock.Anything, examplePlan.ID).Return(examplePlan, nil)
		planDataManager.On("UpdateAccountSubscriptionPlan", mock.Anything, mock.AnythingOfType("*types.AccountSubscriptionPlan")).Return(errors.New("blah"))
		s.planDataManager = planDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything, mock.Anything)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		req = req.WithContext(context.WithValue(req.Context(), updateMiddlewareCtxKey, exampleInput))

		s.UpdateHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, planDataManager, ed)
	})
}

func TestPlansService_ArchiveHandler(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()
	sessionInfoFetcher := func(_ *http.Request) (*types.SessionInfo, error) {
		return exampleUser.ToSessionInfo(), nil
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		examplePlan := fakes.BuildFakePlan()
		s.planIDFetcher = func(req *http.Request) uint64 {
			return examplePlan.ID
		}

		planDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		planDataManager.On("ArchiveAccountSubscriptionPlan", mock.Anything, examplePlan.ID).Return(nil)
		s.planDataManager = planDataManager

		auditLog := &mocktypes.AuditLogEntryDataManager{}
		auditLog.On("AccountSubscriptionLogPlanArchiveEvent", mock.Anything, exampleUser.ID, examplePlan.ID)
		s.auditLog = auditLog

		mc := &mockmetrics.UnitCounter{}
		mc.On("Decrement", mock.Anything).Return()
		s.planCounter = mc

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ArchiveHandler(res, req)

		assert.Equal(t, http.StatusNoContent, res.Code)

		mock.AssertExpectationsForObjects(t, planDataManager, mc)
	})

	T.Run("with no plan in database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		examplePlan := fakes.BuildFakePlan()
		s.planIDFetcher = func(req *http.Request) uint64 {
			return examplePlan.ID
		}

		planDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		planDataManager.On("ArchiveAccountSubscriptionPlan", mock.Anything, examplePlan.ID).Return(sql.ErrNoRows)
		s.planDataManager = planDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeNotFoundResponse", mock.Anything, mock.Anything)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ArchiveHandler(res, req)

		assert.Equal(t, http.StatusNotFound, res.Code)

		mock.AssertExpectationsForObjects(t, planDataManager, ed)
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		examplePlan := fakes.BuildFakePlan()
		s.planIDFetcher = func(req *http.Request) uint64 {
			return examplePlan.ID
		}

		planDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		planDataManager.On("ArchiveAccountSubscriptionPlan", mock.Anything, examplePlan.ID).Return(errors.New("blah"))
		s.planDataManager = planDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything, mock.Anything)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ArchiveHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, planDataManager, ed)
	})

	T.Run("with error removing from search index", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		examplePlan := fakes.BuildFakePlan()
		s.planIDFetcher = func(req *http.Request) uint64 {
			return examplePlan.ID
		}

		planDataManager := &mocktypes.AccountSubscriptionPlanDataManager{}
		planDataManager.On("ArchiveAccountSubscriptionPlan", mock.Anything, examplePlan.ID).Return(nil)
		s.planDataManager = planDataManager

		auditLog := &mocktypes.AuditLogEntryDataManager{}
		auditLog.On("AccountSubscriptionLogPlanArchiveEvent", mock.Anything, exampleUser.ID, examplePlan.ID)
		s.auditLog = auditLog

		mc := &mockmetrics.UnitCounter{}
		mc.On("Decrement", mock.Anything).Return()
		s.planCounter = mc

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ArchiveHandler(res, req)

		assert.Equal(t, http.StatusNoContent, res.Code)

		mock.AssertExpectationsForObjects(t, planDataManager, mc)
	})
}
