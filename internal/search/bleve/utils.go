package bleve

import (
	"fmt"
	"regexp"
)

var (
	belongsToAccountWithMandatedRestrictionRegexp    = regexp.MustCompile(`\+belongsToAccount:[0-9a-zA-Z]+`)
	belongsToAccountWithoutMandatedRestrictionRegexp = regexp.MustCompile(`belongsToAccount:[0-9a-zA-Z]+`)
)

// ensureQueryIsRestrictedToAccount takes a query and userID and ensures that query
// asks that results be restricted to a given user.
func ensureQueryIsRestrictedToAccount(query, accountID string) string {
	switch {
	case belongsToAccountWithMandatedRestrictionRegexp.MatchString(query):
		return query
	case belongsToAccountWithoutMandatedRestrictionRegexp.MatchString(query):
		query = belongsToAccountWithoutMandatedRestrictionRegexp.ReplaceAllString(query, fmt.Sprintf("+belongsToAccount:%s", accountID))
	case !belongsToAccountWithMandatedRestrictionRegexp.MatchString(query):
		query = fmt.Sprintf("%s +belongsToAccount:%s", query, accountID)
	}

	return query
}
