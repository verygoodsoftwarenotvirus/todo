package queries

type BleveQuery struct {
	terms    []string
	prefixes map[string]struct{}
	suffixes map[string]struct{}
}
