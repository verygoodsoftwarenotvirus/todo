package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestQueryLimiter_BuildQueryLimits(t *testing.T) {
	ql, err := NewQueryLimiter("created_on", "updated_on")
	assert.NoError(t, err)
	expected := " WHERE created_on >= ? AND created_on <= ? AND updated_on >= ? AND updated_on <= ? ORDER BY desc LIMIT 50 OFFSET 6100"
	actual := ql.BuildQueryLimits(nil, &QueryFilter{
		Page:          123,
		Limit:         48,
		CreatedAfter:  uint64(time.Now().Unix()),
		CreatedBefore: uint64(time.Now().Unix()),
		UpdatedAfter:  uint64(time.Now().Unix()),
		UpdatedBefore: uint64(time.Now().Unix()),
		SortBy:        SortDescending,
	})

	assert.Equal(t, expected, actual, "expected and actual queries should match")
}
