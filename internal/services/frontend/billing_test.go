package frontend

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/capitalism"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/panicking"
	testutils "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_renderPrice(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		inputsAndExpectations := map[uint32]string{
			12345: "$123.45",
			42069: "$420.69",
			666:   "$6.66",
		}

		for input, expectation := range inputsAndExpectations {
			assert.Equal(t, expectation, renderPrice(input))
		}
	})

	T.Run("with too large a number", func(t *testing.T) {
		t.Parallel()

		fakePanicker := panicking.NewMockPanicker()
		fakePanicker.On("Panic", mock.IsType("")).Return()
		pricePanicker = fakePanicker

		renderPrice(arbitraryPriceMax + 1)

		mock.AssertExpectationsForObjects(t, fakePanicker)
	})
}

func Test_service_handleCheckoutSessionStart(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)
		examplePlanID := "example_plan"
		exampleSessionID := "example_session"

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/whatever?plan=%s", examplePlanID), nil)

		mpm := &capitalism.MockPaymentManager{}
		mpm.On("CreateCheckoutSession", testutils.ContextMatcher, examplePlanID).Return(exampleSessionID, nil)
		s.service.paymentManager = mpm

		s.service.handleCheckoutSessionStart(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
		mock.AssertExpectationsForObjects(t, mpm)
	})

	T.Run("with missing plan ID", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/whatever", nil)

		s.service.handleCheckoutSessionStart(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	T.Run("with error creating checkout session", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)
		examplePlanID := "example_plan"

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/whatever?plan=%s", examplePlanID), nil)

		mpm := &capitalism.MockPaymentManager{}
		mpm.On("CreateCheckoutSession", testutils.ContextMatcher, examplePlanID).Return("", errors.New("blah"))
		s.service.paymentManager = mpm

		s.service.handleCheckoutSessionStart(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)
		mock.AssertExpectationsForObjects(t, mpm)
	})
}

func Test_service_handleCheckoutSuccess(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/whatever", nil)

		s.service.handleCheckoutSuccess(res, req)

		assert.Equal(t, http.StatusTooEarly, res.Code)
	})
}

func Test_service_handleCheckoutCancel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/whatever", nil)

		s.service.handleCheckoutCancel(res, req)

		assert.Equal(t, http.StatusTooEarly, res.Code)
	})
}

func Test_service_handleCheckoutFailure(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/whatever", nil)

		s.service.handleCheckoutFailure(res, req)

		assert.Equal(t, http.StatusTooEarly, res.Code)
	})
}
