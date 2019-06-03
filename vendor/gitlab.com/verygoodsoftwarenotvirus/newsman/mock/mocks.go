package mock

import (
	"gitlab.com/verygoodsoftwarenotvirus/newsman"

	"github.com/stretchr/testify/mock"
)

var _ newsman.Reporter = (*Reporter)(nil)

// Reporter implements newsman's reporter interface
type Reporter struct {
	mock.Mock
}

// Report implements the reporter interface
func (m *Reporter) Report(event newsman.Event) {
	m.Called(event)
}
