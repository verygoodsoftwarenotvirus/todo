package auth

import (
	"testing"
)

func TestProvideWebsocketAuthFunc(t *testing.T) {
	t.Parallel()
	// this is an obligatory test for coverage's sake
	ProvideWebsocketAuthFunc(buildTestService(t))
}
