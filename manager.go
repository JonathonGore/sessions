package sessions

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	// Length of the generated Session ID.
	SessionIDLength = 32
)

type Manager struct {
	cookieName  string   // Name of the cookie we are storing in the users http cookies
	sessionMap  sync.Map // Thread safe map for storing our sessions in memory
	maxLifetime int64    // Expiry time for our sessions
	db          StorageDriver
}

// NewSMManager creates a new session manager based on the given paramaters
func NewManager(cookieName string, maxlifetime int64, db StorageDriver) (*Manager, error) {
	m := &Manager{
		cookieName:  cookieName,
		maxLifetime: maxlifetime,
		sessionMap:  sync.Map{},
		db:          db,
	}

	return m, nil
}

// unwrapSession converts the value stored in memory for a particular session
// into a session object, and reports an error if session is corrupt.
func (m *Manager) unwrapSession(sid string, obj interface{}) (Session, error) {
	var s Session

	s, ok := obj.(Session)
	if !ok {
		return s, fmt.Errorf("corrupt value stored in session map for id: %v", sid)
	}

	return s, nil
}

// GetSession consumes an http request and retrieves the session attached to it
// if available.
func (m *Manager) GetSession(r *http.Request) (Session, error) {
	var s Session

	if !m.HasSession(r) {
		return s, errors.New("no session cookie in http request")
	}

	// Error can be ignored as this is checked in m.HasSession(r)
	cookie, _ := r.Cookie(m.cookieName)

	// Get the session ID attached to the requests cookie.
	sid, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		return s, fmt.Errorf("corrupt value store as session id: %v", err)
	}

	// Attempt to find the session in the in memory cache.
	obj, ok := m.sessionMap.Load(sid)
	if ok {
		s, err = m.unwrapSession(sid, obj)
		if err != nil {
			m.sessionMap.Delete(sid)
			m.db.DeleteSession(sid)
			return s, err
		}
		return s, nil
	}

	// If session map is not found in cache we must consult the db
	s, err = m.db.GetSession(sid)
	if err != nil {
		return s, errors.New("unable to get session, likely invalid session id")
	}

	// Now that we have the session from the db store it in our session map
	m.sessionMap.Store(sid, s)

	return s, nil
}

// HasSession determines if there is a session cookie attached to the request
func (m *Manager) HasSession(r *http.Request) bool {
	cookie, err := r.Cookie(m.cookieName)
	return (err == nil && cookie.Value != "")
}

// SessionStart checks the existence of any sessions related to the current
// request, or creates a new session if none is found.
func (m *Manager) SessionStart(w http.ResponseWriter, r *http.Request, username string) (Session, error) {
	// TODO: Right now if there is a corrupt value for the cookie it will never be repaired
	if m.HasSession(r) {
		return m.GetSession(r)
	}

	sid := generateSessionID()
	s := Session{SID: sid, Username: username, ExpiresOn: time.Now().Add(time.Duration(m.maxLifetime) * time.Second)}

	m.sessionMap.Store(sid, s)
	m.db.InsertSession(s)

	// HTTP only makes it so the cookie is only accessible when sending an http
	// request (so not in javascript)
	cookie := http.Cookie{
		Name:     m.cookieName,
		Value:    url.QueryEscape(sid),
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int(m.maxLifetime),
	}

	http.SetCookie(w, &cookie)

	return s, nil
}

// SessionDestroy removes the session stored in the requests cookies.
// Typically called on logout.
func (m *Manager) SessionDestroy(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie(m.cookieName)
	if err != nil || cookie.Value == "" {
		return nil
	}

	// Remove session from cache and database
	m.sessionMap.Delete(cookie.Value)
	m.db.DeleteSession(cookie.Value)

	// Overwrite the current cookie with an expired one
	ec := http.Cookie{Name: m.cookieName, Path: "/", HttpOnly: true, Expires: time.Unix(0, 0), MaxAge: -1}
	http.SetCookie(w, &ec)

	return nil
}

// GenerateSessionID produces a unique sessionID for a new session.
func generateSessionID() string {
	b := make([]byte, SessionIDLength)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}
