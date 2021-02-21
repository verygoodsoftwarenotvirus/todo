package bleve

import (
	"fmt"
	"regexp"
)

var (
	belongsToUserWithMandatedRestrictionRegexp    = regexp.MustCompile(`\+belongsToAccount:\d+`)
	belongsToUserWithoutMandatedRestrictionRegexp = regexp.MustCompile(`belongsToAccount:\d+`)
)

// ensureQueryIsRestrictedToUser takes a query and userID and ensures that query
// asks that results be restricted to a given user.
func ensureQueryIsRestrictedToUser(query string, userID uint64) string {
	switch {
	case belongsToUserWithMandatedRestrictionRegexp.MatchString(query):
		return query
	case belongsToUserWithoutMandatedRestrictionRegexp.MatchString(query):
		query = belongsToUserWithoutMandatedRestrictionRegexp.ReplaceAllString(query, fmt.Sprintf("+belongsToAccount:%d", userID))
	case !belongsToUserWithMandatedRestrictionRegexp.MatchString(query):
		query = fmt.Sprintf("%s +belongsToAccount:%d", query, userID)
	}

	return query
}
