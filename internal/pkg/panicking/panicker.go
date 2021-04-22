package panicking

// Panicker abstracts panic for our tests and such.
type Panicker interface {
	Panic(interface{})
	Panicf(format string, args ...interface{})
}
