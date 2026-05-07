package testutil

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/Devlaner/devlane/api/internal/model"
	"github.com/Devlaner/devlane/api/internal/store"
	"gorm.io/gorm"
)

// LoginAs creates a fresh session row for the given user and returns the
// 40-char hex session key. Pass the key as the last argument to ts.GET /
// POST / PATCH / PUT / DELETE (or ts.Do) to authenticate the request via
// the session_id cookie.
//
// This bypasses bcrypt+SignIn for speed but exercises the same persistence
// path that production auth.Service.createSession uses (via SessionStore).
func LoginAs(t testing.TB, db *gorm.DB, user *model.User) string {
	t.Helper()
	if user == nil {
		t.Fatal("LoginAs: user is nil")
	}
	var raw [20]byte
	if _, err := rand.Read(raw[:]); err != nil {
		t.Fatalf("LoginAs: rand: %v", err)
	}
	key := hex.EncodeToString(raw[:])
	if err := store.NewSessionStore(db).Create(context.Background(), key, user.ID); err != nil {
		t.Fatalf("LoginAs: create session: %v", err)
	}
	return key
}
