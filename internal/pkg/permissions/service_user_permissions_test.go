package permissions

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceUserPermissions(t *testing.T) {
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
}
