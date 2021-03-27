package bleve

import (
	"fmt"
	"regexp"
)

var (
	belongsToAccountWithMandatedRestrictionRegexp    = regexp.MustCompile(`\+belongsToAccount:\d+`)
	belongsToAccountWithoutMandatedRestrictionRegexp = regexp.MustCompile(`belongsToAccount:\d+`)
)

// ensureQueryIsRestrictedToUser takes a query and userID and ensures that query
// asks that results be restricted to a given user.
func ensureQueryIsRestrictedToUser(query string, userID uint64) string {
	switch {
	case belongsToAccountWithMandatedRestrictionRegexp.MatchString(query):
		return query
	case belongsToAccountWithoutMandatedRestrictionRegexp.MatchString(query):
		query = belongsToAccountWithoutMandatedRestrictionRegexp.ReplaceAllString(query, fmt.Sprintf("+belongsToAccount:%d", userID))
	case !belongsToAccountWithMandatedRestrictionRegexp.MatchString(query):
		query = fmt.Sprintf("%s +belongsToAccount:%d", query, userID)
	}

	return query
}
