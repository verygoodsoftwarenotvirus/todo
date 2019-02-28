package resources

// Reporter is our reporting interface for a particular resource
type Reporter interface {
	Add()
	Subtract()
	ResourceCount() uint64
}
