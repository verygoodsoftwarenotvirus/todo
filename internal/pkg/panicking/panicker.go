package panicking

import "fmt"

// Panicker abstracts panic for our tests and such.
type Panicker interface {
	Panic(msg string)
	Panicf(format string, args ...interface{})
}

// NewProductionPanicker produces a production-ready panicker that will actually panic when called.
func NewProductionPanicker() Panicker {
	return stdLibPanicker{}
}

type stdLibPanicker struct{}

func (p stdLibPanicker) Panic(msg string) {
	panic(msg)
}

func (p stdLibPanicker) Panicf(format string, args ...interface{}) {
	p.Panic(fmt.Sprintf(format, args...))
}
