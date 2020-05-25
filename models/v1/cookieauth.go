package models

// CookieAuth represents what we encode in our authentication cookies.
type CookieAuth struct {
	UserID   uint64 `json:"-"`
	Admin    bool   `json:"-"`
	Username string `json:"-"`
}

// StatusResponse is what we encode when the frontend wants to check auth status
type StatusResponse struct {
	Authenticated bool `json:"isAuthenticated"`
	IsAdmin       bool `json:"isAdmin"`
}
