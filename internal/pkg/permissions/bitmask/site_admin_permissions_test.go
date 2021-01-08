package bitmask

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSiteAdminPermissions_CanCycleCookieSecrets(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := NewPermissionBitmask(0)
		assert.False(t, x.CanCycleCookieSecrets())

		y := NewPermissionBitmask(0 | uint32(cycleCookieSecretPermission))
		assert.True(t, y.CanCycleCookieSecrets())
	})
}

func TestSiteAdminPermissions_CanBanUsers(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := SiteAdminPermissions(0)
		assert.False(t, x.CanBanUsers())

		y := SiteAdminPermissions(0 | uint32(banUserPermission))
		assert.True(t, y.CanBanUsers())
	})
}

func TestSiteAdminPermissions_CanTerminateAccounts(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := SiteAdminPermissions(0)
		assert.False(t, x.CanTerminateAccounts())

		y := SiteAdminPermissions(0 | uint32(canTerminateAccountsPermission))
		assert.True(t, y.CanTerminateAccounts())
	})
}

func TestSiteAdminPermissions_hasReservedUnusedPermission4(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := SiteAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission4())

		y := SiteAdminPermissions(0 | uint32(unusedSiteAdminPermission4))
		assert.True(t, y.hasReservedUnusedPermission4())
	})
}

func TestSiteAdminPermissions_hasReservedUnusedPermission5(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := SiteAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission5())

		y := SiteAdminPermissions(0 | uint32(unusedSiteAdminPermission5))
		assert.True(t, y.hasReservedUnusedPermission5())
	})
}

func TestSiteAdminPermissions_hasReservedUnusedPermission6(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := SiteAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission6())

		y := SiteAdminPermissions(0 | uint32(unusedSiteAdminPermission6))
		assert.True(t, y.hasReservedUnusedPermission6())
	})
}

func TestSiteAdminPermissions_hasReservedUnusedPermission7(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := SiteAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission7())

		y := SiteAdminPermissions(0 | uint32(unusedSiteAdminPermission7))
		assert.True(t, y.hasReservedUnusedPermission7())
	})
}

func TestSiteAdminPermissions_hasReservedUnusedPermission8(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := SiteAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission8())

		y := SiteAdminPermissions(0 | uint32(unusedSiteAdminPermission8))
		assert.True(t, y.hasReservedUnusedPermission8())
	})
}

func TestSiteAdminPermissions_hasReservedUnusedPermission9(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := SiteAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission9())

		y := SiteAdminPermissions(0 | uint32(unusedSiteAdminPermission9))
		assert.True(t, y.hasReservedUnusedPermission9())
	})
}

func TestSiteAdminPermissions_hasReservedUnusedPermission10(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := SiteAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission10())

		y := SiteAdminPermissions(0 | uint32(unusedSiteAdminPermission10))
		assert.True(t, y.hasReservedUnusedPermission10())
	})
}

func TestSiteAdminPermissions_hasReservedUnusedPermission11(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := SiteAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission11())

		y := SiteAdminPermissions(0 | uint32(unusedSiteAdminPermission11))
		assert.True(t, y.hasReservedUnusedPermission11())
	})
}

func TestSiteAdminPermissions_hasReservedUnusedPermission12(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := SiteAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission12())

		y := SiteAdminPermissions(0 | uint32(unusedSiteAdminPermission12))
		assert.True(t, y.hasReservedUnusedPermission12())
	})
}

func TestSiteAdminPermissions_hasReservedUnusedPermission13(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := SiteAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission13())

		y := SiteAdminPermissions(0 | uint32(unusedSiteAdminPermission13))
		assert.True(t, y.hasReservedUnusedPermission13())
	})
}

func TestSiteAdminPermissions_hasReservedUnusedPermission14(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := SiteAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission14())

		y := SiteAdminPermissions(0 | uint32(unusedSiteAdminPermission14))
		assert.True(t, y.hasReservedUnusedPermission14())
	})
}

func TestSiteAdminPermissions_hasReservedUnusedPermission15(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := SiteAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission15())

		y := SiteAdminPermissions(0 | uint32(unusedSiteAdminPermission15))
		assert.True(t, y.hasReservedUnusedPermission15())
	})
}

func TestSiteAdminPermissions_hasReservedUnusedPermission16(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := SiteAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission16())

		y := SiteAdminPermissions(0 | uint32(unusedSiteAdminPermission16))
		assert.True(t, y.hasReservedUnusedPermission16())
	})
}

func TestSiteAdminPermissions_hasReservedUnusedPermission17(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := SiteAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission17())

		y := SiteAdminPermissions(0 | uint32(unusedSiteAdminPermission17))
		assert.True(t, y.hasReservedUnusedPermission17())
	})
}

func TestSiteAdminPermissions_hasReservedUnusedPermission18(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := SiteAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission18())

		y := SiteAdminPermissions(0 | uint32(unusedSiteAdminPermission18))
		assert.True(t, y.hasReservedUnusedPermission18())
	})
}

func TestSiteAdminPermissions_hasReservedUnusedPermission19(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := SiteAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission19())

		y := SiteAdminPermissions(0 | uint32(unusedSiteAdminPermission19))
		assert.True(t, y.hasReservedUnusedPermission19())
	})
}

func TestSiteAdminPermissions_hasReservedUnusedPermission20(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := SiteAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission20())

		y := SiteAdminPermissions(0 | uint32(unusedSiteAdminPermission20))
		assert.True(t, y.hasReservedUnusedPermission20())
	})
}

func TestSiteAdminPermissions_hasReservedUnusedPermission21(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := SiteAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission21())

		y := SiteAdminPermissions(0 | uint32(unusedSiteAdminPermission21))
		assert.True(t, y.hasReservedUnusedPermission21())
	})
}

func TestSiteAdminPermissions_hasReservedUnusedPermission22(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := SiteAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission22())

		y := SiteAdminPermissions(0 | uint32(unusedSiteAdminPermission22))
		assert.True(t, y.hasReservedUnusedPermission22())
	})
}

func TestSiteAdminPermissions_hasReservedUnusedPermission23(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := SiteAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission23())

		y := SiteAdminPermissions(0 | uint32(unusedSiteAdminPermission23))
		assert.True(t, y.hasReservedUnusedPermission23())
	})
}

func TestSiteAdminPermissions_hasReservedUnusedPermission24(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := SiteAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission24())

		y := SiteAdminPermissions(0 | uint32(unusedSiteAdminPermission24))
		assert.True(t, y.hasReservedUnusedPermission24())
	})
}

func TestSiteAdminPermissions_hasReservedUnusedPermission25(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := SiteAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission25())

		y := SiteAdminPermissions(0 | uint32(unusedSiteAdminPermission25))
		assert.True(t, y.hasReservedUnusedPermission25())
	})
}

func TestSiteAdminPermissions_hasReservedUnusedPermission26(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := SiteAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission26())

		y := SiteAdminPermissions(0 | uint32(unusedSiteAdminPermission26))
		assert.True(t, y.hasReservedUnusedPermission26())
	})
}

func TestSiteAdminPermissions_hasReservedUnusedPermission27(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := SiteAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission27())

		y := SiteAdminPermissions(0 | uint32(unusedSiteAdminPermission27))
		assert.True(t, y.hasReservedUnusedPermission27())
	})
}

func TestSiteAdminPermissions_hasReservedUnusedPermission28(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := SiteAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission28())

		y := SiteAdminPermissions(0 | uint32(unusedSiteAdminPermission28))
		assert.True(t, y.hasReservedUnusedPermission28())
	})
}

func TestSiteAdminPermissions_hasReservedUnusedPermission29(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := SiteAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission29())

		y := SiteAdminPermissions(0 | uint32(unusedSiteAdminPermission29))
		assert.True(t, y.hasReservedUnusedPermission29())
	})
}

func TestSiteAdminPermissions_hasReservedUnusedPermission30(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := SiteAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission30())

		y := SiteAdminPermissions(0 | uint32(unusedSiteAdminPermission30))
		assert.True(t, y.hasReservedUnusedPermission30())
	})
}

func TestSiteAdminPermissions_hasReservedUnusedPermission31(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := SiteAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission31())

		y := SiteAdminPermissions(0 | uint32(unusedSiteAdminPermission31))
		assert.True(t, y.hasReservedUnusedPermission31())
	})
}

func TestSiteAdminPermissions_hasReservedUnusedPermission32(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := SiteAdminPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission32())

		y := SiteAdminPermissions(0 | uint32(unusedSiteAdminPermission32))
		assert.True(t, y.hasReservedUnusedPermission32())
	})
}
