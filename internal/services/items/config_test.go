package items

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &Config{
			PreWritesTopicName:   "blah",
			PreUpdatesTopicName:  "blah",
			PreArchivesTopicName: "blah",
			SearchIndexPath:      "blah",
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})
}
