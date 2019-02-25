package sessions

import "time"

// Session defines the structure of the session object stored by Sessions.
type Session struct {
	ID        string `json:"id"`
	Values    map[string]interface{}
	CreatedAt time.Time `json:"created-at"`
	ExpiresAt time.Time `json:"expires-at"`
}
