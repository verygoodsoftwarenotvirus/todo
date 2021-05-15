package requests

import (
	"net/http"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
)

func TestBuilder_BuildGetAuditLogEntryRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/admin/audit_log/%d"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAuditLogEntry := fakes.BuildFakeAuditLogEntry()
		spec := newRequestSpec(true, http.MethodGet, "", expectedPath, exampleAuditLogEntry.ID)

		actual, err := helper.builder.BuildGetAuditLogEntryRequest(helper.ctx, exampleAuditLogEntry.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})
}

func TestBuilder_BuildGetAuditLogEntriesRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/admin/audit_log"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		filter := types.DefaultQueryFilter()
		spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

		actual, err := helper.builder.BuildGetAuditLogEntriesRequest(helper.ctx, filter)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})
}