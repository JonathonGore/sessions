package sessions

import "time"

// Session defines the structure of the session object stored by Sessions.
type Session struct {
	SID       string    `json:"sid"`
	Username  string    `json:"username"`
	CreatedOn time.Time `json:"created-on"`
	ExpiresOn time.Time `json:"expires-on"`
}
