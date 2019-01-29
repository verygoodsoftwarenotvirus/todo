package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBuildQueryLimits(t *testing.T) {
	expected := " WHERE created_on >= ? AND created_on <= ? AND updated_on >= ? AND updated_on <= ? ORDER BY desc LIMIT 50 OFFSET 6100"
	actual := BuildQueryLimits(
		nil,
		&QueryFilter{
			Page:          123,
			Limit:         48,
			CreatedAfter:  uint64(time.Now().Unix()),
			CreatedBefore: uint64(time.Now().Unix()),
			UpdatedAfter:  uint64(time.Now().Unix()),
			UpdatedBefore: uint64(time.Now().Unix()),
			SortBy:        SortDescending,
		},
		"created_on",
		"updated_on",
	)

	assert.Equal(t, expected, actual, "expected and actual queries should match")
}
