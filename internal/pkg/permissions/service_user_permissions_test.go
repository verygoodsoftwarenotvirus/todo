package permissions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccountUserPermissions_CanCreateItems(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := NewServiceUserPermissions(0)
		assert.False(t, x.CanCreateItems())

		y := NewServiceUserPermissions(0 | uint32(CanCreateItems))
		assert.True(t, y.CanCreateItems())
	})
}

func TestAccountUserPermissions_CanUpdateItems(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.CanUpdateItems())

		y := ServiceUserPermissions(0 | uint32(CanUpdateItems))
		assert.True(t, y.CanUpdateItems())
	})
}

func TestAccountUserPermissions_CanArchiveItems(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.CanArchiveItems())

		y := ServiceUserPermissions(0 | uint32(CanArchiveItems))
		assert.True(t, y.CanArchiveItems())
	})
}

func TestAccountUserPermissions_hasReservedUnusedPermission13(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission13())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission13))
		assert.True(t, y.hasReservedUnusedPermission13())
	})
}

func TestAccountUserPermissions_hasReservedUnusedPermission14(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission14())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission14))
		assert.True(t, y.hasReservedUnusedPermission14())
	})
}

func TestAccountUserPermissions_hasReservedUnusedPermission15(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission15())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission15))
		assert.True(t, y.hasReservedUnusedPermission15())
	})
}

func TestAccountUserPermissions_hasReservedUnusedPermission16(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission16())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission16))
		assert.True(t, y.hasReservedUnusedPermission16())
	})
}

func TestAccountUserPermissions_hasReservedUnusedPermission17(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission17())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission17))
		assert.True(t, y.hasReservedUnusedPermission17())
	})
}

func TestAccountUserPermissions_hasReservedUnusedPermission18(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission18())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission18))
		assert.True(t, y.hasReservedUnusedPermission18())
	})
}

func TestAccountUserPermissions_hasReservedUnusedPermission19(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission19())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission19))
		assert.True(t, y.hasReservedUnusedPermission19())
	})
}

func TestAccountUserPermissions_hasReservedUnusedPermission20(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission20())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission20))
		assert.True(t, y.hasReservedUnusedPermission20())
	})
}

func TestAccountUserPermissions_hasReservedUnusedPermission21(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission21())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission21))
		assert.True(t, y.hasReservedUnusedPermission21())
	})
}

func TestAccountUserPermissions_hasReservedUnusedPermission22(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission22())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission22))
		assert.True(t, y.hasReservedUnusedPermission22())
	})
}

func TestAccountUserPermissions_hasReservedUnusedPermission23(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission23())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission23))
		assert.True(t, y.hasReservedUnusedPermission23())
	})
}

func TestAccountUserPermissions_hasReservedUnusedPermission24(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission24())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission24))
		assert.True(t, y.hasReservedUnusedPermission24())
	})
}

func TestAccountUserPermissions_hasReservedUnusedPermission25(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission25())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission25))
		assert.True(t, y.hasReservedUnusedPermission25())
	})
}

func TestAccountUserPermissions_hasReservedUnusedPermission26(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission26())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission26))
		assert.True(t, y.hasReservedUnusedPermission26())
	})
}

func TestAccountUserPermissions_hasReservedUnusedPermission27(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission27())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission27))
		assert.True(t, y.hasReservedUnusedPermission27())
	})
}

func TestAccountUserPermissions_hasReservedUnusedPermission28(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission28())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission28))
		assert.True(t, y.hasReservedUnusedPermission28())
	})
}

func TestAccountUserPermissions_hasReservedUnusedPermission29(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission29())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission29))
		assert.True(t, y.hasReservedUnusedPermission29())
	})
}

func TestAccountUserPermissions_hasReservedUnusedPermission30(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission30())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission30))
		assert.True(t, y.hasReservedUnusedPermission30())
	})
}

func TestAccountUserPermissions_hasReservedUnusedPermission31(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission31())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission31))
		assert.True(t, y.hasReservedUnusedPermission31())
	})
}

func TestAccountUserPermissions_hasReservedUnusedPermission32(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission32())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission32))
		assert.True(t, y.hasReservedUnusedPermission32())
	})
}
