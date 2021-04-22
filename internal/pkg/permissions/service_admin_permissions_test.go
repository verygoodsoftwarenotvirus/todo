package permissions

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServiceAdminPermission(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, NewServiceAdminPermissions(0))
	})
}

func TestServiceAdminPermission_Value(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermission(0)

		actual, err := x.Value()
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})
}

func TestServiceAdminPermission_Scan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermission(0)
		expected := int64(123)

		assert.NoError(t, x.Scan(expected))
		assert.Equal(t, ServiceAdminPermission(expected), x)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermission(0)

		assert.NoError(t, x.Scan("123"))
	})
}

func TestServiceAdminPermission_MarshalJSON(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermission(0)

		actual, err := x.MarshalJSON()
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})
}

func TestServiceAdminPermission_UnmarshalJSON(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermission(0)

		assert.NoError(t, x.UnmarshalJSON([]byte("0")))
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermission(0)

		assert.Error(t, x.UnmarshalJSON([]byte("/")))
	})
}

func TestServiceAdminPermission_Summary(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermission(math.MaxInt64)

		assert.NotNil(t, x.Summary())
	})
}

func TestServiceAdminPermission_IsServiceAdmin(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermission(math.MaxInt64)

		assert.True(t, x.IsServiceAdmin())
	})
}

func TestServiceAdminPermissions(T *testing.T) {
	T.Parallel()

	T.Run("with no permissions", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermission(0)

		assert.False(t, x.CanCycleCookieSecrets())
		assert.False(t, x.CanBanUsers())
		assert.False(t, x.CanTerminateAccounts())
		assert.False(t, x.CanImpersonateAccounts())
		assert.False(t, x.hasUnusedServiceAdminPermission5())
		assert.False(t, x.hasUnusedServiceAdminPermission6())
		assert.False(t, x.hasUnusedServiceAdminPermission7())
		assert.False(t, x.hasUnusedServiceAdminPermission8())
		assert.False(t, x.hasUnusedServiceAdminPermission9())
		assert.False(t, x.hasUnusedServiceAdminPermission10())
		assert.False(t, x.hasUnusedServiceAdminPermission11())
		assert.False(t, x.hasUnusedServiceAdminPermission12())
		assert.False(t, x.hasUnusedServiceAdminPermission13())
		assert.False(t, x.hasUnusedServiceAdminPermission14())
		assert.False(t, x.hasUnusedServiceAdminPermission15())
		assert.False(t, x.hasUnusedServiceAdminPermission16())
		assert.False(t, x.hasUnusedServiceAdminPermission17())
		assert.False(t, x.hasUnusedServiceAdminPermission18())
		assert.False(t, x.hasUnusedServiceAdminPermission19())
		assert.False(t, x.hasUnusedServiceAdminPermission20())
		assert.False(t, x.hasUnusedServiceAdminPermission21())
		assert.False(t, x.hasUnusedServiceAdminPermission22())
		assert.False(t, x.hasUnusedServiceAdminPermission23())
		assert.False(t, x.hasUnusedServiceAdminPermission24())
		assert.False(t, x.hasUnusedServiceAdminPermission25())
		assert.False(t, x.hasUnusedServiceAdminPermission26())
		assert.False(t, x.hasUnusedServiceAdminPermission27())
		assert.False(t, x.hasUnusedServiceAdminPermission28())
		assert.False(t, x.hasUnusedServiceAdminPermission29())
		assert.False(t, x.hasUnusedServiceAdminPermission30())
		assert.False(t, x.hasUnusedServiceAdminPermission31())
		assert.False(t, x.hasUnusedServiceAdminPermission32())
		assert.False(t, x.hasUnusedServiceAdminPermission33())
		assert.False(t, x.hasUnusedServiceAdminPermission34())
		assert.False(t, x.hasUnusedServiceAdminPermission35())
		assert.False(t, x.hasUnusedServiceAdminPermission36())
		assert.False(t, x.hasUnusedServiceAdminPermission37())
		assert.False(t, x.hasUnusedServiceAdminPermission38())
		assert.False(t, x.hasUnusedServiceAdminPermission39())
		assert.False(t, x.hasUnusedServiceAdminPermission40())
		assert.False(t, x.hasUnusedServiceAdminPermission41())
		assert.False(t, x.hasUnusedServiceAdminPermission42())
		assert.False(t, x.hasUnusedServiceAdminPermission43())
		assert.False(t, x.hasUnusedServiceAdminPermission44())
		assert.False(t, x.hasUnusedServiceAdminPermission45())
		assert.False(t, x.hasUnusedServiceAdminPermission46())
		assert.False(t, x.hasUnusedServiceAdminPermission47())
		assert.False(t, x.hasUnusedServiceAdminPermission48())
		assert.False(t, x.hasUnusedServiceAdminPermission49())
		assert.False(t, x.hasUnusedServiceAdminPermission50())
		assert.False(t, x.hasUnusedServiceAdminPermission51())
		assert.False(t, x.hasUnusedServiceAdminPermission52())
		assert.False(t, x.hasUnusedServiceAdminPermission53())
		assert.False(t, x.hasUnusedServiceAdminPermission54())
		assert.False(t, x.hasUnusedServiceAdminPermission55())
		assert.False(t, x.hasUnusedServiceAdminPermission56())
		assert.False(t, x.hasUnusedServiceAdminPermission57())
		assert.False(t, x.hasUnusedServiceAdminPermission58())
		assert.False(t, x.hasUnusedServiceAdminPermission59())
		assert.False(t, x.hasUnusedServiceAdminPermission61())
		assert.False(t, x.hasUnusedServiceAdminPermission62())
		assert.False(t, x.hasUnusedServiceAdminPermission63())
		assert.False(t, x.hasUnusedServiceAdminPermission64())
	})

	T.Run("with all permissions", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermission(math.MaxInt64)

		assert.True(t, x.CanCycleCookieSecrets())
		assert.True(t, x.CanBanUsers())
		assert.True(t, x.CanTerminateAccounts())
		assert.True(t, x.CanImpersonateAccounts())
		assert.True(t, x.hasUnusedServiceAdminPermission5())
		assert.True(t, x.hasUnusedServiceAdminPermission6())
		assert.True(t, x.hasUnusedServiceAdminPermission7())
		assert.True(t, x.hasUnusedServiceAdminPermission8())
		assert.True(t, x.hasUnusedServiceAdminPermission9())
		assert.True(t, x.hasUnusedServiceAdminPermission10())
		assert.True(t, x.hasUnusedServiceAdminPermission11())
		assert.True(t, x.hasUnusedServiceAdminPermission12())
		assert.True(t, x.hasUnusedServiceAdminPermission13())
		assert.True(t, x.hasUnusedServiceAdminPermission14())
		assert.True(t, x.hasUnusedServiceAdminPermission15())
		assert.True(t, x.hasUnusedServiceAdminPermission16())
		assert.True(t, x.hasUnusedServiceAdminPermission17())
		assert.True(t, x.hasUnusedServiceAdminPermission18())
		assert.True(t, x.hasUnusedServiceAdminPermission19())
		assert.True(t, x.hasUnusedServiceAdminPermission20())
		assert.True(t, x.hasUnusedServiceAdminPermission21())
		assert.True(t, x.hasUnusedServiceAdminPermission22())
		assert.True(t, x.hasUnusedServiceAdminPermission23())
		assert.True(t, x.hasUnusedServiceAdminPermission24())
		assert.True(t, x.hasUnusedServiceAdminPermission25())
		assert.True(t, x.hasUnusedServiceAdminPermission26())
		assert.True(t, x.hasUnusedServiceAdminPermission27())
		assert.True(t, x.hasUnusedServiceAdminPermission28())
		assert.True(t, x.hasUnusedServiceAdminPermission29())
		assert.True(t, x.hasUnusedServiceAdminPermission30())
		assert.True(t, x.hasUnusedServiceAdminPermission31())
		assert.True(t, x.hasUnusedServiceAdminPermission32())
		assert.True(t, x.hasUnusedServiceAdminPermission33())
		assert.True(t, x.hasUnusedServiceAdminPermission34())
		assert.True(t, x.hasUnusedServiceAdminPermission35())
		assert.True(t, x.hasUnusedServiceAdminPermission36())
		assert.True(t, x.hasUnusedServiceAdminPermission37())
		assert.True(t, x.hasUnusedServiceAdminPermission38())
		assert.True(t, x.hasUnusedServiceAdminPermission39())
		assert.True(t, x.hasUnusedServiceAdminPermission40())
		assert.True(t, x.hasUnusedServiceAdminPermission41())
		assert.True(t, x.hasUnusedServiceAdminPermission42())
		assert.True(t, x.hasUnusedServiceAdminPermission43())
		assert.True(t, x.hasUnusedServiceAdminPermission44())
		assert.True(t, x.hasUnusedServiceAdminPermission45())
		assert.True(t, x.hasUnusedServiceAdminPermission46())
		assert.True(t, x.hasUnusedServiceAdminPermission47())
		assert.True(t, x.hasUnusedServiceAdminPermission48())
		assert.True(t, x.hasUnusedServiceAdminPermission49())
		assert.True(t, x.hasUnusedServiceAdminPermission50())
		assert.True(t, x.hasUnusedServiceAdminPermission51())
		assert.True(t, x.hasUnusedServiceAdminPermission52())
		assert.True(t, x.hasUnusedServiceAdminPermission53())
		assert.True(t, x.hasUnusedServiceAdminPermission54())
		assert.True(t, x.hasUnusedServiceAdminPermission55())
		assert.True(t, x.hasUnusedServiceAdminPermission56())
		assert.True(t, x.hasUnusedServiceAdminPermission57())
		assert.True(t, x.hasUnusedServiceAdminPermission58())
		assert.True(t, x.hasUnusedServiceAdminPermission59())
		assert.True(t, x.hasUnusedServiceAdminPermission61())
		assert.True(t, x.hasUnusedServiceAdminPermission62())
		assert.True(t, x.hasUnusedServiceAdminPermission63())
		assert.True(t, x.hasUnusedServiceAdminPermission64())
	})
}
