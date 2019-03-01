package resources

// Reporter is our reporting interface for a particular resource
type Reporter interface {
	Inc()
	Add(value uint64)
	Subtract()
	ResourceCount() uint64
}
