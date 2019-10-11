package metrics

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegisterDefaultViews(t *testing.T) {
	t.Parallel()
	// obligatory
	require.NoError(t, RegisterDefaultViews())
}
