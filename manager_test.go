package sessions

import (
	"testing"
)

func TestGenerateSessionID(t *testing.T) {
	sid := generateSessionID()
	if len(sid) < SessionIDLength {
		t.Errorf("Expected generated session id to be at least length: %v - found: %v", SessionIDLength, len(sid))
	}
}
