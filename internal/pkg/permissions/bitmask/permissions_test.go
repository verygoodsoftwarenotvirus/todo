package bitmask

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_permissionMask_CanCycleCookieSecrets(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := NewPermissionBitmask(0)
		assert.False(t, x.CanCycleCookieSecrets())

		y := NewPermissionBitmask(0 | uint32(cycleCookieSecretPermission))
		assert.True(t, y.CanCycleCookieSecrets())
	})
}

func Test_permissionMask_hasReservedUnusedPermission2(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := AdminPermissionsBitmask(0)
		assert.False(t, x.hasReservedUnusedPermission2())

		y := AdminPermissionsBitmask(0 | uint32(reservedUnusedPermission2))
		assert.True(t, y.hasReservedUnusedPermission2())
	})
}

func Test_permissionMask_hasReservedUnusedPermission3(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := AdminPermissionsBitmask(0)
		assert.False(t, x.hasReservedUnusedPermission3())

		y := AdminPermissionsBitmask(0 | uint32(reservedUnusedPermission3))
		assert.True(t, y.hasReservedUnusedPermission3())
	})
}

func Test_permissionMask_hasReservedUnusedPermission4(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := AdminPermissionsBitmask(0)
		assert.False(t, x.hasReservedUnusedPermission4())

		y := AdminPermissionsBitmask(0 | uint32(reservedUnusedPermission4))
		assert.True(t, y.hasReservedUnusedPermission4())
	})
}

func Test_permissionMask_hasReservedUnusedPermission5(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := AdminPermissionsBitmask(0)
		assert.False(t, x.hasReservedUnusedPermission5())

		y := AdminPermissionsBitmask(0 | uint32(reservedUnusedPermission5))
		assert.True(t, y.hasReservedUnusedPermission5())
	})
}

func Test_permissionMask_hasReservedUnusedPermission6(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := AdminPermissionsBitmask(0)
		assert.False(t, x.hasReservedUnusedPermission6())

		y := AdminPermissionsBitmask(0 | uint32(reservedUnusedPermission6))
		assert.True(t, y.hasReservedUnusedPermission6())
	})
}

func Test_permissionMask_hasReservedUnusedPermission7(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := AdminPermissionsBitmask(0)
		assert.False(t, x.hasReservedUnusedPermission7())

		y := AdminPermissionsBitmask(0 | uint32(reservedUnusedPermission7))
		assert.True(t, y.hasReservedUnusedPermission7())
	})
}

func Test_permissionMask_hasReservedUnusedPermission8(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := AdminPermissionsBitmask(0)
		assert.False(t, x.hasReservedUnusedPermission8())

		y := AdminPermissionsBitmask(0 | uint32(reservedUnusedPermission8))
		assert.True(t, y.hasReservedUnusedPermission8())
	})
}

func Test_permissionMask_hasReservedUnusedPermission9(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := AdminPermissionsBitmask(0)
		assert.False(t, x.hasReservedUnusedPermission9())

		y := AdminPermissionsBitmask(0 | uint32(reservedUnusedPermission9))
		assert.True(t, y.hasReservedUnusedPermission9())
	})
}

func Test_permissionMask_hasReservedUnusedPermission10(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := AdminPermissionsBitmask(0)
		assert.False(t, x.hasReservedUnusedPermission10())

		y := AdminPermissionsBitmask(0 | uint32(reservedUnusedPermission10))
		assert.True(t, y.hasReservedUnusedPermission10())
	})
}

func Test_permissionMask_hasReservedUnusedPermission11(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := AdminPermissionsBitmask(0)
		assert.False(t, x.hasReservedUnusedPermission11())

		y := AdminPermissionsBitmask(0 | uint32(reservedUnusedPermission11))
		assert.True(t, y.hasReservedUnusedPermission11())
	})
}

func Test_permissionMask_hasReservedUnusedPermission12(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := AdminPermissionsBitmask(0)
		assert.False(t, x.hasReservedUnusedPermission12())

		y := AdminPermissionsBitmask(0 | uint32(reservedUnusedPermission12))
		assert.True(t, y.hasReservedUnusedPermission12())
	})
}

func Test_permissionMask_hasReservedUnusedPermission13(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := AdminPermissionsBitmask(0)
		assert.False(t, x.hasReservedUnusedPermission13())

		y := AdminPermissionsBitmask(0 | uint32(reservedUnusedPermission13))
		assert.True(t, y.hasReservedUnusedPermission13())
	})
}

func Test_permissionMask_hasReservedUnusedPermission14(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := AdminPermissionsBitmask(0)
		assert.False(t, x.hasReservedUnusedPermission14())

		y := AdminPermissionsBitmask(0 | uint32(reservedUnusedPermission14))
		assert.True(t, y.hasReservedUnusedPermission14())
	})
}

func Test_permissionMask_hasReservedUnusedPermission15(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := AdminPermissionsBitmask(0)
		assert.False(t, x.hasReservedUnusedPermission15())

		y := AdminPermissionsBitmask(0 | uint32(reservedUnusedPermission15))
		assert.True(t, y.hasReservedUnusedPermission15())
	})
}

func Test_permissionMask_hasReservedUnusedPermission16(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := AdminPermissionsBitmask(0)
		assert.False(t, x.hasReservedUnusedPermission16())

		y := AdminPermissionsBitmask(0 | uint32(reservedUnusedPermission16))
		assert.True(t, y.hasReservedUnusedPermission16())
	})
}

func Test_permissionMask_hasReservedUnusedPermission17(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := AdminPermissionsBitmask(0)
		assert.False(t, x.hasReservedUnusedPermission17())

		y := AdminPermissionsBitmask(0 | uint32(reservedUnusedPermission17))
		assert.True(t, y.hasReservedUnusedPermission17())
	})
}

func Test_permissionMask_hasReservedUnusedPermission18(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := AdminPermissionsBitmask(0)
		assert.False(t, x.hasReservedUnusedPermission18())

		y := AdminPermissionsBitmask(0 | uint32(reservedUnusedPermission18))
		assert.True(t, y.hasReservedUnusedPermission18())
	})
}

func Test_permissionMask_hasReservedUnusedPermission19(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := AdminPermissionsBitmask(0)
		assert.False(t, x.hasReservedUnusedPermission19())

		y := AdminPermissionsBitmask(0 | uint32(reservedUnusedPermission19))
		assert.True(t, y.hasReservedUnusedPermission19())
	})
}

func Test_permissionMask_hasReservedUnusedPermission20(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := AdminPermissionsBitmask(0)
		assert.False(t, x.hasReservedUnusedPermission20())

		y := AdminPermissionsBitmask(0 | uint32(reservedUnusedPermission20))
		assert.True(t, y.hasReservedUnusedPermission20())
	})
}

func Test_permissionMask_hasReservedUnusedPermission21(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := AdminPermissionsBitmask(0)
		assert.False(t, x.hasReservedUnusedPermission21())

		y := AdminPermissionsBitmask(0 | uint32(reservedUnusedPermission21))
		assert.True(t, y.hasReservedUnusedPermission21())
	})
}

func Test_permissionMask_hasReservedUnusedPermission22(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := AdminPermissionsBitmask(0)
		assert.False(t, x.hasReservedUnusedPermission22())

		y := AdminPermissionsBitmask(0 | uint32(reservedUnusedPermission22))
		assert.True(t, y.hasReservedUnusedPermission22())
	})
}

func Test_permissionMask_hasReservedUnusedPermission23(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := AdminPermissionsBitmask(0)
		assert.False(t, x.hasReservedUnusedPermission23())

		y := AdminPermissionsBitmask(0 | uint32(reservedUnusedPermission23))
		assert.True(t, y.hasReservedUnusedPermission23())
	})
}

func Test_permissionMask_hasReservedUnusedPermission24(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := AdminPermissionsBitmask(0)
		assert.False(t, x.hasReservedUnusedPermission24())

		y := AdminPermissionsBitmask(0 | uint32(reservedUnusedPermission24))
		assert.True(t, y.hasReservedUnusedPermission24())
	})
}

func Test_permissionMask_hasReservedUnusedPermission25(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := AdminPermissionsBitmask(0)
		assert.False(t, x.hasReservedUnusedPermission25())

		y := AdminPermissionsBitmask(0 | uint32(reservedUnusedPermission25))
		assert.True(t, y.hasReservedUnusedPermission25())
	})
}

func Test_permissionMask_hasReservedUnusedPermission26(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := AdminPermissionsBitmask(0)
		assert.False(t, x.hasReservedUnusedPermission26())

		y := AdminPermissionsBitmask(0 | uint32(reservedUnusedPermission26))
		assert.True(t, y.hasReservedUnusedPermission26())
	})
}

func Test_permissionMask_hasReservedUnusedPermission27(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := AdminPermissionsBitmask(0)
		assert.False(t, x.hasReservedUnusedPermission27())

		y := AdminPermissionsBitmask(0 | uint32(reservedUnusedPermission27))
		assert.True(t, y.hasReservedUnusedPermission27())
	})
}

func Test_permissionMask_hasReservedUnusedPermission28(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := AdminPermissionsBitmask(0)
		assert.False(t, x.hasReservedUnusedPermission28())

		y := AdminPermissionsBitmask(0 | uint32(reservedUnusedPermission28))
		assert.True(t, y.hasReservedUnusedPermission28())
	})
}

func Test_permissionMask_hasReservedUnusedPermission29(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := AdminPermissionsBitmask(0)
		assert.False(t, x.hasReservedUnusedPermission29())

		y := AdminPermissionsBitmask(0 | uint32(reservedUnusedPermission29))
		assert.True(t, y.hasReservedUnusedPermission29())
	})
}

func Test_permissionMask_hasReservedUnusedPermission30(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := AdminPermissionsBitmask(0)
		assert.False(t, x.hasReservedUnusedPermission30())

		y := AdminPermissionsBitmask(0 | uint32(reservedUnusedPermission30))
		assert.True(t, y.hasReservedUnusedPermission30())
	})
}

func Test_permissionMask_hasReservedUnusedPermission31(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := AdminPermissionsBitmask(0)
		assert.False(t, x.hasReservedUnusedPermission31())

		y := AdminPermissionsBitmask(0 | uint32(reservedUnusedPermission31))
		assert.True(t, y.hasReservedUnusedPermission31())
	})
}

func Test_permissionMask_hasReservedUnusedPermission32(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		x := AdminPermissionsBitmask(0)
		assert.False(t, x.hasReservedUnusedPermission32())

		y := AdminPermissionsBitmask(0 | uint32(reservedUnusedPermission32))
		assert.True(t, y.hasReservedUnusedPermission32())
	})
}
