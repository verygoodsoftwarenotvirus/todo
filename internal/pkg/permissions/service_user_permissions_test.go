package permissions

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServiceUserPermission(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, NewServiceUserPermissions(0))
	})
}

func TestServiceUserPermission_Summary(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermission(math.MaxInt64)

		assert.NotNil(t, x.Summary())
	})
}

func TestServiceUserPermission_Value(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermission(0)

		actual, err := x.Value()
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})
}

func TestServiceUserPermission_Scan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermission(0)
		expected := int64(123)

		assert.NoError(t, x.Scan(expected))
		assert.Equal(t, ServiceUserPermission(expected), x)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermission(0)

		assert.NoError(t, x.Scan("123"))
	})
}

func TestServiceUserPermission_MarshalJSON(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermission(0)

		actual, err := x.MarshalJSON()
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})
}

func TestServiceUserPermission_UnmarshalJSON(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermission(0)

		assert.NoError(t, x.UnmarshalJSON([]byte("0")))
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermission(0)

		assert.Error(t, x.UnmarshalJSON([]byte("/")))
	})
}

func TestServiceUserPermissions(T *testing.T) {
	T.Parallel()

	T.Run("with no permissions", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermission(0)

		assert.False(t, x.CanManageWebhooks())
		assert.False(t, x.CanManageAPIClients())
		assert.False(t, x.hasUnusedAccountUserPermission3())
		assert.False(t, x.hasUnusedAccountUserPermission4())
		assert.False(t, x.hasUnusedAccountUserPermission5())
		assert.False(t, x.hasUnusedAccountUserPermission6())
		assert.False(t, x.hasUnusedAccountUserPermission7())
		assert.False(t, x.hasUnusedAccountUserPermission8())
		assert.False(t, x.hasUnusedAccountUserPermission9())
		assert.False(t, x.hasUnusedAccountUserPermission10())
		assert.False(t, x.hasUnusedAccountUserPermission11())
		assert.False(t, x.hasUnusedAccountUserPermission12())
		assert.False(t, x.hasUnusedAccountUserPermission13())
		assert.False(t, x.hasUnusedAccountUserPermission14())
		assert.False(t, x.hasUnusedAccountUserPermission15())
		assert.False(t, x.hasUnusedAccountUserPermission16())
		assert.False(t, x.hasUnusedAccountUserPermission17())
		assert.False(t, x.hasUnusedAccountUserPermission18())
		assert.False(t, x.hasUnusedAccountUserPermission19())
		assert.False(t, x.hasUnusedAccountUserPermission20())
		assert.False(t, x.hasUnusedAccountUserPermission21())
		assert.False(t, x.hasUnusedAccountUserPermission22())
		assert.False(t, x.hasUnusedAccountUserPermission23())
		assert.False(t, x.hasUnusedAccountUserPermission24())
		assert.False(t, x.hasUnusedAccountUserPermission25())
		assert.False(t, x.hasUnusedAccountUserPermission26())
		assert.False(t, x.hasUnusedAccountUserPermission27())
		assert.False(t, x.hasUnusedAccountUserPermission28())
		assert.False(t, x.hasUnusedAccountUserPermission29())
		assert.False(t, x.hasUnusedAccountUserPermission30())
		assert.False(t, x.hasUnusedAccountUserPermission31())
		assert.False(t, x.hasUnusedAccountUserPermission32())
		assert.False(t, x.hasUnusedAccountUserPermission33())
		assert.False(t, x.hasUnusedAccountUserPermission34())
		assert.False(t, x.hasUnusedAccountUserPermission35())
		assert.False(t, x.hasUnusedAccountUserPermission36())
		assert.False(t, x.hasUnusedAccountUserPermission37())
		assert.False(t, x.hasUnusedAccountUserPermission38())
		assert.False(t, x.hasUnusedAccountUserPermission39())
		assert.False(t, x.hasUnusedAccountUserPermission40())
		assert.False(t, x.hasUnusedAccountUserPermission41())
		assert.False(t, x.hasUnusedAccountUserPermission42())
		assert.False(t, x.hasUnusedAccountUserPermission43())
		assert.False(t, x.hasUnusedAccountUserPermission44())
		assert.False(t, x.hasUnusedAccountUserPermission45())
		assert.False(t, x.hasUnusedAccountUserPermission46())
		assert.False(t, x.hasUnusedAccountUserPermission47())
		assert.False(t, x.hasUnusedAccountUserPermission48())
		assert.False(t, x.hasUnusedAccountUserPermission49())
		assert.False(t, x.hasUnusedAccountUserPermission50())
		assert.False(t, x.hasUnusedAccountUserPermission51())
		assert.False(t, x.hasUnusedAccountUserPermission52())
		assert.False(t, x.hasUnusedAccountUserPermission53())
		assert.False(t, x.hasUnusedAccountUserPermission54())
		assert.False(t, x.hasUnusedAccountUserPermission55())
		assert.False(t, x.hasUnusedAccountUserPermission56())
		assert.False(t, x.hasUnusedAccountUserPermission57())
		assert.False(t, x.hasUnusedAccountUserPermission58())
		assert.False(t, x.hasUnusedAccountUserPermission59())
		assert.False(t, x.hasUnusedAccountUserPermission61())
		assert.False(t, x.hasUnusedAccountUserPermission62())
		assert.False(t, x.hasUnusedAccountUserPermission63())
		assert.False(t, x.hasUnusedAccountUserPermission64())
	})

	T.Run("with all permissions", func(t *testing.T) {
		t.Parallel()
		x := ServiceUserPermission(math.MaxInt64)

		assert.True(t, x.CanManageWebhooks())
		assert.True(t, x.CanManageAPIClients())
		assert.True(t, x.hasUnusedAccountUserPermission3())
		assert.True(t, x.hasUnusedAccountUserPermission4())
		assert.True(t, x.hasUnusedAccountUserPermission5())
		assert.True(t, x.hasUnusedAccountUserPermission6())
		assert.True(t, x.hasUnusedAccountUserPermission7())
		assert.True(t, x.hasUnusedAccountUserPermission8())
		assert.True(t, x.hasUnusedAccountUserPermission9())
		assert.True(t, x.hasUnusedAccountUserPermission10())
		assert.True(t, x.hasUnusedAccountUserPermission11())
		assert.True(t, x.hasUnusedAccountUserPermission12())
		assert.True(t, x.hasUnusedAccountUserPermission13())
		assert.True(t, x.hasUnusedAccountUserPermission14())
		assert.True(t, x.hasUnusedAccountUserPermission15())
		assert.True(t, x.hasUnusedAccountUserPermission16())
		assert.True(t, x.hasUnusedAccountUserPermission17())
		assert.True(t, x.hasUnusedAccountUserPermission18())
		assert.True(t, x.hasUnusedAccountUserPermission19())
		assert.True(t, x.hasUnusedAccountUserPermission20())
		assert.True(t, x.hasUnusedAccountUserPermission21())
		assert.True(t, x.hasUnusedAccountUserPermission22())
		assert.True(t, x.hasUnusedAccountUserPermission23())
		assert.True(t, x.hasUnusedAccountUserPermission24())
		assert.True(t, x.hasUnusedAccountUserPermission25())
		assert.True(t, x.hasUnusedAccountUserPermission26())
		assert.True(t, x.hasUnusedAccountUserPermission27())
		assert.True(t, x.hasUnusedAccountUserPermission28())
		assert.True(t, x.hasUnusedAccountUserPermission29())
		assert.True(t, x.hasUnusedAccountUserPermission30())
		assert.True(t, x.hasUnusedAccountUserPermission31())
		assert.True(t, x.hasUnusedAccountUserPermission32())
		assert.True(t, x.hasUnusedAccountUserPermission33())
		assert.True(t, x.hasUnusedAccountUserPermission34())
		assert.True(t, x.hasUnusedAccountUserPermission35())
		assert.True(t, x.hasUnusedAccountUserPermission36())
		assert.True(t, x.hasUnusedAccountUserPermission37())
		assert.True(t, x.hasUnusedAccountUserPermission38())
		assert.True(t, x.hasUnusedAccountUserPermission39())
		assert.True(t, x.hasUnusedAccountUserPermission40())
		assert.True(t, x.hasUnusedAccountUserPermission41())
		assert.True(t, x.hasUnusedAccountUserPermission42())
		assert.True(t, x.hasUnusedAccountUserPermission43())
		assert.True(t, x.hasUnusedAccountUserPermission44())
		assert.True(t, x.hasUnusedAccountUserPermission45())
		assert.True(t, x.hasUnusedAccountUserPermission46())
		assert.True(t, x.hasUnusedAccountUserPermission47())
		assert.True(t, x.hasUnusedAccountUserPermission48())
		assert.True(t, x.hasUnusedAccountUserPermission49())
		assert.True(t, x.hasUnusedAccountUserPermission50())
		assert.True(t, x.hasUnusedAccountUserPermission51())
		assert.True(t, x.hasUnusedAccountUserPermission52())
		assert.True(t, x.hasUnusedAccountUserPermission53())
		assert.True(t, x.hasUnusedAccountUserPermission54())
		assert.True(t, x.hasUnusedAccountUserPermission55())
		assert.True(t, x.hasUnusedAccountUserPermission56())
		assert.True(t, x.hasUnusedAccountUserPermission57())
		assert.True(t, x.hasUnusedAccountUserPermission58())
		assert.True(t, x.hasUnusedAccountUserPermission59())
		assert.True(t, x.hasUnusedAccountUserPermission61())
		assert.True(t, x.hasUnusedAccountUserPermission62())
		assert.True(t, x.hasUnusedAccountUserPermission63())
		assert.True(t, x.hasUnusedAccountUserPermission64())
	})
}
