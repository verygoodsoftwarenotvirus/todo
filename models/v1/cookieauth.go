package models

// CookieAuth represents what we encode in our cookie
type CookieAuth struct {
	UserID   uint64
	Admin    bool
	Username string
}
