package sessions

import "time"

// Session defines the structure of the session object stored by Sessions.
type Session struct {
	ID        string `json:"id"`
	Values    map[string]interface{}
	CreatedOn time.Time `json:"created-on"`
	ExpiresOn time.Time `json:"expires-on"`
}
