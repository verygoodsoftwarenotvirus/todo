package models

import (
	"bytes"
	"encoding/gob"
)

const (
	// SessionInfoKey is the non-string type we use for referencing SessionInfo structs
	SessionInfoKey ContextKey = "session_info"
)

func init() {
	gob.Register(&SessionInfo{})
}

type (
	// SessionInfo represents what we encode in our authentication cookies.
	SessionInfo struct {
		UserID      uint64 `json:"-"`
		UserIsAdmin bool   `json:"-"`
	}

	// StatusResponse is what we encode when the frontend wants to check auth status
	StatusResponse struct {
		Authenticated bool `json:"isAuthenticated"`
		IsAdmin       bool `json:"isAdmin"`
	}
)

// ToBytes returns the gob encoded session info
func (i *SessionInfo) ToBytes() []byte {
	var b bytes.Buffer

	if err := gob.NewEncoder(&b).Encode(i); err != nil {
		panic(err)
	}

	return b.Bytes()
}
