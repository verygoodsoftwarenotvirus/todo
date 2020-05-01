package models

// CookieAuth represents what we encode in our authentication cookies.
type CookieAuth struct {
	UserID   uint64
	Admin    bool
	Username string
}
