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

// Add implements our reporter's Add method
func (r *Reporter) Add() {
	atomic.AddUint64(&r.count, 1)
}

// Subtract implements our reporter's Subtract method
func (r *Reporter) Subtract() {
	atomic.AddUint64(&r.count, ^uint64(0))
}

// ResourceCount implements our reporter's ResourceCount method
func (r *Reporter) ResourceCount() uint64 {
	return r.count
}
