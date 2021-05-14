package panicking

import "fmt"

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
