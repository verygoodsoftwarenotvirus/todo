package authorization

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func hasPermission(p Permission, roles ...string) bool {
	for _, r := range roles {
		if !globalAuthorizer.IsGranted(r, p, nil) {
			return false
		}
	}

	return true
}

// CanCreateItems returns whether a user can create items or not.
func CanCreateItems(roles ...string) bool {
	return hasPermission(CreateItemsPermission, roles...)
}

// CanSeeItems returns whether a user can view items or not.
func CanSeeItems(roles ...string) bool {
	return hasPermission(ReadItemsPermission, roles...)
}

// CanSearchItems returns whether a user can search items or not.
func CanSearchItems(roles ...string) bool {
	return hasPermission(SearchItemsPermission, roles...)
}

// CanUpdateItems returns whether a user can update items or not.
func CanUpdateItems(roles ...string) bool {
	return hasPermission(UpdateItemsPermission, roles...)
}

// CanDeleteItems returns whether a user can delete items or not.
func CanDeleteItems(roles ...string) bool {
	return hasPermission(ArchiveItemsPermission, roles...)
}
