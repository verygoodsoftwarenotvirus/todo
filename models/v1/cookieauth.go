package models

// CookieAuth represents what we encode in our cookie
type CookieAuth struct {
	Admin    bool
	Username string
}
