package noop

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/resources/v1"
)

// ProvideNoopResourceReporter does what it says on the tin
func ProvideNoopResourceReporter() resources.Reporter {
	return &Reporter{}
}

// Reporter is an in-memory reporter
type Reporter struct{}

// Inc implements our reporter's Inc requirement
func (r *Reporter) Inc() {

}

// Add implements our reporter's Add method
func (r *Reporter) Add(value uint64) {

}

// Subtract implements our reporter's Subtract method
func (r *Reporter) Subtract() {

}

// ResourceCount implements our reporter's ResourceCount method
func (r *Reporter) ResourceCount() uint64 {
	return 0
}
