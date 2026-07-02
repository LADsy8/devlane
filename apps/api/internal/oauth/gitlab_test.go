package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newGitLabTestServer(t *testing.T, body map[string]interface{}) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v4/user" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET request, got %s", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer tok" {
			t.Errorf("expected bearer token auth header, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(body)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestGitLabGetUserInfo_RejectsUnconfirmedEmail(t *testing.T) {
	srv := newGitLabTestServer(t, map[string]interface{}{
		"id":           42,
		"email":        "victim@example.com",
		"name":         "Attacker",
		"avatar_url":   "https://example.com/a.png",
		"confirmed_at": nil,
	})

	provider := NewGitLabProvider(ProviderConfig{}, srv.URL)
	_, err := provider.GetUserInfo(context.Background(), &TokenData{AccessToken: "tok"})
	if !errors.Is(err, ErrEmailNotVerified) {
		t.Fatalf("expected ErrEmailNotVerified, got %v", err)
	}
}

func TestGitLabGetUserInfo_AcceptsConfirmedEmail(t *testing.T) {
	srv := newGitLabTestServer(t, map[string]interface{}{
		"id":           42,
		"email":        "user@example.com",
		"name":         "Real User",
		"avatar_url":   "https://example.com/a.png",
		"confirmed_at": "2024-01-01T00:00:00.000Z",
	})

	provider := NewGitLabProvider(ProviderConfig{}, srv.URL)
	info, err := provider.GetUserInfo(context.Background(), &TokenData{AccessToken: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Email != "user@example.com" {
		t.Fatalf("expected email to be returned, got %q", info.Email)
	}
	if info.ProviderID != "42" {
		t.Fatalf("expected provider id 42, got %q", info.ProviderID)
	}
}
