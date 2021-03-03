package permissions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceAdminPermissions_CanCycleCookieSecrets(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := NewServiceAdminPermissions(0)
		assert.False(t, x.CanCycleCookieSecrets())

		y := NewServiceAdminPermissions(0 | uint32(cycleCookieSecretPermission))
		assert.True(t, y.CanCycleCookieSecrets())
	})
}

func TestServiceAdminPermissions_CanBanUsers(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermissions(0)
		assert.False(t, x.CanBanUsers())

		y := ServiceAdminPermissions(0 | uint32(banUserPermission))
		assert.True(t, y.CanBanUsers())
	})
}

func TestServiceAdminPermissions_CanTerminateAccounts(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermissions(0)
		assert.False(t, x.CanTerminateAccounts())

		y := ServiceAdminPermissions(0 | uint32(canTerminateAccountsPermission))
		assert.True(t, y.CanTerminateAccounts())
	})
}

func TestServiceAdminPermissions_hasReservedUnusedPermission4(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission4())

		y := ServiceAdminPermissions(0 | uint32(unusedServiceAdminPermission4))
		assert.True(t, y.hasReservedUnusedPermission4())
	})
}

func TestServiceAdminPermissions_hasReservedUnusedPermission5(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission5())

		y := ServiceAdminPermissions(0 | uint32(unusedServiceAdminPermission5))
		assert.True(t, y.hasReservedUnusedPermission5())
	})
}

func TestServiceAdminPermissions_hasReservedUnusedPermission6(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission6())

		y := ServiceAdminPermissions(0 | uint32(unusedServiceAdminPermission6))
		assert.True(t, y.hasReservedUnusedPermission6())
	})
}

func TestServiceAdminPermissions_hasReservedUnusedPermission7(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission7())

		y := ServiceAdminPermissions(0 | uint32(unusedServiceAdminPermission7))
		assert.True(t, y.hasReservedUnusedPermission7())
	})
}

func TestServiceAdminPermissions_hasReservedUnusedPermission8(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission8())

		y := ServiceAdminPermissions(0 | uint32(unusedServiceAdminPermission8))
		assert.True(t, y.hasReservedUnusedPermission8())
	})
}

func TestServiceAdminPermissions_hasReservedUnusedPermission9(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission9())

		y := ServiceAdminPermissions(0 | uint32(unusedServiceAdminPermission9))
		assert.True(t, y.hasReservedUnusedPermission9())
	})
}

func TestServiceAdminPermissions_hasReservedUnusedPermission10(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission10())

		y := ServiceAdminPermissions(0 | uint32(unusedServiceAdminPermission10))
		assert.True(t, y.hasReservedUnusedPermission10())
	})
}

func TestServiceAdminPermissions_hasReservedUnusedPermission11(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission11())

		y := ServiceAdminPermissions(0 | uint32(unusedServiceAdminPermission11))
		assert.True(t, y.hasReservedUnusedPermission11())
	})
}

func TestServiceAdminPermissions_hasReservedUnusedPermission12(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission12())

		y := ServiceAdminPermissions(0 | uint32(unusedServiceAdminPermission12))
		assert.True(t, y.hasReservedUnusedPermission12())
	})
}

func TestServiceAdminPermissions_hasReservedUnusedPermission13(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission13())

		y := ServiceAdminPermissions(0 | uint32(unusedServiceAdminPermission13))
		assert.True(t, y.hasReservedUnusedPermission13())
	})
}

func TestServiceAdminPermissions_hasReservedUnusedPermission14(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission14())

		y := ServiceAdminPermissions(0 | uint32(unusedServiceAdminPermission14))
		assert.True(t, y.hasReservedUnusedPermission14())
	})
}

func TestServiceAdminPermissions_hasReservedUnusedPermission15(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission15())

		y := ServiceAdminPermissions(0 | uint32(unusedServiceAdminPermission15))
		assert.True(t, y.hasReservedUnusedPermission15())
	})
}

func TestServiceAdminPermissions_hasReservedUnusedPermission16(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission16())

		y := ServiceAdminPermissions(0 | uint32(unusedServiceAdminPermission16))
		assert.True(t, y.hasReservedUnusedPermission16())
	})
}

func TestServiceAdminPermissions_hasReservedUnusedPermission17(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission17())

		y := ServiceAdminPermissions(0 | uint32(unusedServiceAdminPermission17))
		assert.True(t, y.hasReservedUnusedPermission17())
	})
}

func TestServiceAdminPermissions_hasReservedUnusedPermission18(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission18())

		y := ServiceAdminPermissions(0 | uint32(unusedServiceAdminPermission18))
		assert.True(t, y.hasReservedUnusedPermission18())
	})
}

func TestServiceAdminPermissions_hasReservedUnusedPermission19(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission19())

		y := ServiceAdminPermissions(0 | uint32(unusedServiceAdminPermission19))
		assert.True(t, y.hasReservedUnusedPermission19())
	})
}

func TestServiceAdminPermissions_hasReservedUnusedPermission20(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission20())

		y := ServiceAdminPermissions(0 | uint32(unusedServiceAdminPermission20))
		assert.True(t, y.hasReservedUnusedPermission20())
	})
}

func TestServiceAdminPermissions_hasReservedUnusedPermission21(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission21())

		y := ServiceAdminPermissions(0 | uint32(unusedServiceAdminPermission21))
		assert.True(t, y.hasReservedUnusedPermission21())
	})
}

func TestServiceAdminPermissions_hasReservedUnusedPermission22(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission22())

		y := ServiceAdminPermissions(0 | uint32(unusedServiceAdminPermission22))
		assert.True(t, y.hasReservedUnusedPermission22())
	})
}

func TestServiceAdminPermissions_hasReservedUnusedPermission23(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission23())

		y := ServiceAdminPermissions(0 | uint32(unusedServiceAdminPermission23))
		assert.True(t, y.hasReservedUnusedPermission23())
	})
}

func TestServiceAdminPermissions_hasReservedUnusedPermission24(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission24())

		y := ServiceAdminPermissions(0 | uint32(unusedServiceAdminPermission24))
		assert.True(t, y.hasReservedUnusedPermission24())
	})
}

func TestServiceAdminPermissions_hasReservedUnusedPermission25(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission25())

		y := ServiceAdminPermissions(0 | uint32(unusedServiceAdminPermission25))
		assert.True(t, y.hasReservedUnusedPermission25())
	})
}

func TestServiceAdminPermissions_hasReservedUnusedPermission26(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission26())

		y := ServiceAdminPermissions(0 | uint32(unusedServiceAdminPermission26))
		assert.True(t, y.hasReservedUnusedPermission26())
	})
}

func TestServiceAdminPermissions_hasReservedUnusedPermission27(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission27())

		y := ServiceAdminPermissions(0 | uint32(unusedServiceAdminPermission27))
		assert.True(t, y.hasReservedUnusedPermission27())
	})
}

func TestServiceAdminPermissions_hasReservedUnusedPermission28(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission28())

		y := ServiceAdminPermissions(0 | uint32(unusedServiceAdminPermission28))
		assert.True(t, y.hasReservedUnusedPermission28())
	})
}

func TestServiceAdminPermissions_hasReservedUnusedPermission29(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission29())

		y := ServiceAdminPermissions(0 | uint32(unusedServiceAdminPermission29))
		assert.True(t, y.hasReservedUnusedPermission29())
	})
}

func TestServiceAdminPermissions_hasReservedUnusedPermission30(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission30())

		y := ServiceAdminPermissions(0 | uint32(unusedServiceAdminPermission30))
		assert.True(t, y.hasReservedUnusedPermission30())
	})
}

func TestServiceAdminPermissions_hasReservedUnusedPermission31(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission31())

		y := ServiceAdminPermissions(0 | uint32(unusedServiceAdminPermission31))
		assert.True(t, y.hasReservedUnusedPermission31())
	})
}

func TestServiceAdminPermissions_hasReservedUnusedPermission32(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission32())

		y := ServiceAdminPermissions(0 | uint32(unusedServiceAdminPermission32))
		assert.True(t, y.hasReservedUnusedPermission32())
	})
}
