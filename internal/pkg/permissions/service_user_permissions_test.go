package permissions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceUserPermissions_hasReservedUnusedPermission10(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission10())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission10))
		assert.True(t, y.hasReservedUnusedPermission10())
	})
}

func TestServiceUserPermissions_hasReservedUnusedPermission11(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission11())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission11))
		assert.True(t, y.hasReservedUnusedPermission11())
	})
}

func TestServiceUserPermissions_hasReservedUnusedPermission12(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission12())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission12))
		assert.True(t, y.hasReservedUnusedPermission12())
	})
}

func TestServiceUserPermissions_hasReservedUnusedPermission13(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission13())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission13))
		assert.True(t, y.hasReservedUnusedPermission13())
	})
}

func TestServiceUserPermissions_hasReservedUnusedPermission14(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission14())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission14))
		assert.True(t, y.hasReservedUnusedPermission14())
	})
}

func TestServiceUserPermissions_hasReservedUnusedPermission15(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission15())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission15))
		assert.True(t, y.hasReservedUnusedPermission15())
	})
}

func TestServiceUserPermissions_hasReservedUnusedPermission16(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission16())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission16))
		assert.True(t, y.hasReservedUnusedPermission16())
	})
}

func TestServiceUserPermissions_hasReservedUnusedPermission17(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission17())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission17))
		assert.True(t, y.hasReservedUnusedPermission17())
	})
}

func TestServiceUserPermissions_hasReservedUnusedPermission18(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission18())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission18))
		assert.True(t, y.hasReservedUnusedPermission18())
	})
}

func TestServiceUserPermissions_hasReservedUnusedPermission19(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission19())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission19))
		assert.True(t, y.hasReservedUnusedPermission19())
	})
}

func TestServiceUserPermissions_hasReservedUnusedPermission20(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission20())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission10))
		assert.True(t, y.hasReservedUnusedPermission20())
	})
}

func TestServiceUserPermissions_hasReservedUnusedPermission21(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission21())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission11))
		assert.True(t, y.hasReservedUnusedPermission21())
	})
}

func TestServiceUserPermissions_hasReservedUnusedPermission22(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission22())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission12))
		assert.True(t, y.hasReservedUnusedPermission22())
	})
}

func TestServiceUserPermissions_hasReservedUnusedPermission23(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission23())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission13))
		assert.True(t, y.hasReservedUnusedPermission23())
	})
}

func TestServiceUserPermissions_hasReservedUnusedPermission24(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission24())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission14))
		assert.True(t, y.hasReservedUnusedPermission24())
	})
}

func TestServiceUserPermissions_hasReservedUnusedPermission25(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission25())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission15))
		assert.True(t, y.hasReservedUnusedPermission25())
	})
}

func TestServiceUserPermissions_hasReservedUnusedPermission26(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission26())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission16))
		assert.True(t, y.hasReservedUnusedPermission26())
	})
}

func TestServiceUserPermissions_hasReservedUnusedPermission27(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission27())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission17))
		assert.True(t, y.hasReservedUnusedPermission27())
	})
}

func TestServiceUserPermissions_hasReservedUnusedPermission28(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission28())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission18))
		assert.True(t, y.hasReservedUnusedPermission28())
	})
}

func TestServiceUserPermissions_hasReservedUnusedPermission29(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission29())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission19))
		assert.True(t, y.hasReservedUnusedPermission29())
	})
}

func TestServiceUserPermissions_hasReservedUnusedPermission3(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission3())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission3))
		assert.True(t, y.hasReservedUnusedPermission3())
	})
}

func TestServiceUserPermissions_hasReservedUnusedPermission30(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission30())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission10))
		assert.True(t, y.hasReservedUnusedPermission30())
	})
}

func TestServiceUserPermissions_hasReservedUnusedPermission31(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission31())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission11))
		assert.True(t, y.hasReservedUnusedPermission31())
	})
}

func TestServiceUserPermissions_hasReservedUnusedPermission32(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission32())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission12))
		assert.True(t, y.hasReservedUnusedPermission32())
	})
}

func TestServiceUserPermissions_hasReservedUnusedPermission4(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission4())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission4))
		assert.True(t, y.hasReservedUnusedPermission4())
	})
}

func TestServiceUserPermissions_hasReservedUnusedPermission5(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission5())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission5))
		assert.True(t, y.hasReservedUnusedPermission5())
	})
}

func TestServiceUserPermissions_hasReservedUnusedPermission6(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission6())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission6))
		assert.True(t, y.hasReservedUnusedPermission6())
	})
}

func TestServiceUserPermissions_hasReservedUnusedPermission7(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission7())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission7))
		assert.True(t, y.hasReservedUnusedPermission7())
	})
}

func TestServiceUserPermissions_hasReservedUnusedPermission8(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission8())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission8))
		assert.True(t, y.hasReservedUnusedPermission8())
	})
}

func TestServiceUserPermissions_hasReservedUnusedPermission9(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		x := ServiceUserPermissions(0)
		assert.False(t, x.hasReservedUnusedPermission9())

		y := ServiceUserPermissions(0 | uint32(unusedAccountUserPermission9))
		assert.True(t, y.hasReservedUnusedPermission9())
	})
}
