package memory

import (
	"sync/atomic"

	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/resources/v1"
)

var _ resources.Reporter = (*Reporter)(nil)

// Reporter is an in-memory reporter
type Reporter struct {
	count uint64
}

// Inc implements our reporter's Inc requirement
func (r *Reporter) Inc() {
	r.Add(1)
}

// Add implements our reporter's Add requirement
func (r *Reporter) Add(value uint64) {
	atomic.AddUint64(&r.count, value)
}

// Subtract implements our reporter's Subtract requirement
func (r *Reporter) Subtract() {
	atomic.AddUint64(&r.count, ^uint64(0))
}

// ResourceCount implements our reporter's ResourceCount requirement
func (r *Reporter) ResourceCount() uint64 {
	return r.count
}
