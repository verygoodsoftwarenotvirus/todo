package panicking

import "fmt"

// Panicker abstracts panic for our tests and such.
type Panicker interface {
	Panic(interface{})
	Panicf(format string, args ...interface{})
}

// NewProductionPanicker produces a production-ready panicker that will actually panic when called.
func NewProductionPanicker() Panicker {
	return stdLibPanicker{}
}

type stdLibPanicker struct{}

func (p stdLibPanicker) Panic(msg interface{}) {
	panic(msg)
}

func (p stdLibPanicker) Panicf(format string, args ...interface{}) {
	p.Panic(fmt.Sprintf(format, args...))
}
