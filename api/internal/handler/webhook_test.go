package handler_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"testing"

	devcrypto "github.com/Devlaner/devlane/api/internal/crypto"
	"github.com/Devlaner/devlane/api/internal/model"
	"github.com/Devlaner/devlane/api/internal/store"
	"github.com/Devlaner/devlane/api/internal/testutil"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func signGitHubPayload(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// seedGitHubWebhookSecret stores an encrypted webhook secret in instance_settings
// the way handler/instance.go would when an admin saves the github_app section.
func seedGitHubWebhookSecret(t *testing.T, db *gorm.DB, secret string) {
	t.Helper()
	settings := store.NewInstanceSettingStore(db)
	require.NoError(t, settings.Upsert(context.Background(), "github_app", model.JSONMap{
		"app_id":             "12345",
		"app_name":           "test-app",
		"client_id":          "Iv1.testclient",
		"webhook_secret":     devcrypto.EncryptOrPlain(secret),
		"webhook_secret_set": true,
	}))
}

func TestWebhook_GitHub_MissingSignature(t *testing.T) {
	ts := testutil.NewTestServer(t)
	seedGitHubWebhookSecret(t, ts.DB, "test-webhook-secret")

	body := []byte(`{"action":"opened"}`)
	rr := ts.DoWithHeaders(http.MethodPost, "/webhooks/github", body, http.Header{
		"X-GitHub-Event":    []string{"pull_request"},
		"X-GitHub-Delivery": []string{"deadbeef-1234"},
	})
	require.Equal(t, http.StatusUnauthorized, rr.Code, "body=%s", rr.Body.String())
}

func TestWebhook_GitHub_BadSignature(t *testing.T) {
	ts := testutil.NewTestServer(t)
	seedGitHubWebhookSecret(t, ts.DB, "test-webhook-secret")

	body := []byte(`{"action":"opened"}`)
	rr := ts.DoWithHeaders(http.MethodPost, "/webhooks/github", body, http.Header{
		"X-GitHub-Event":      []string{"pull_request"},
		"X-GitHub-Delivery":   []string{"deadbeef-1234"},
		"X-Hub-Signature-256": []string{"sha256=" + strings.Repeat("0", 64)},
		"Content-Type":        []string{"application/json"},
	})
	require.Equal(t, http.StatusUnauthorized, rr.Code, "body=%s", rr.Body.String())
}

func TestWebhook_GitHub_ValidSignature_NoTrailingSlash(t *testing.T) {
	ts := testutil.NewTestServer(t)
	const secret = "test-webhook-secret"
	seedGitHubWebhookSecret(t, ts.DB, secret)

	body := []byte(`{"action":"opened","number":1}`)
	sig := signGitHubPayload(secret, body)

	rr := ts.DoWithHeaders(http.MethodPost, "/webhooks/github", body, http.Header{
		"X-GitHub-Event":      []string{"pull_request"},
		"X-GitHub-Delivery":   []string{"deadbeef-1234"},
		"X-Hub-Signature-256": []string{sig},
		"Content-Type":        []string{"application/json"},
	})
	require.Equal(t, http.StatusOK, rr.Code, "body=%s", rr.Body.String())
}

func TestWebhook_GitHub_ValidSignature_TrailingSlash(t *testing.T) {
	ts := testutil.NewTestServer(t)
	const secret = "test-webhook-secret"
	seedGitHubWebhookSecret(t, ts.DB, secret)

	body := []byte(`{"action":"opened","number":2}`)
	sig := signGitHubPayload(secret, body)

	rr := ts.DoWithHeaders(http.MethodPost, "/webhooks/github/", body, http.Header{
		"X-GitHub-Event":      []string{"pull_request"},
		"X-GitHub-Delivery":   []string{"deadbeef-5678"},
		"X-Hub-Signature-256": []string{sig},
		"Content-Type":        []string{"application/json"},
	})
	require.Equal(t, http.StatusOK, rr.Code, "body=%s", rr.Body.String())
}

func TestWebhook_GitHub_NoSecretConfigured(t *testing.T) {
	// With no github_app settings, webhook secret is empty → VerifySignature
	// returns "secret is not configured" → 401.
	ts := testutil.NewTestServer(t)

	body := []byte(`{"action":"opened"}`)
	rr := ts.DoWithHeaders(http.MethodPost, "/webhooks/github", body, http.Header{
		"X-GitHub-Event":      []string{"pull_request"},
		"X-GitHub-Delivery":   []string{"any"},
		"X-Hub-Signature-256": []string{"sha256=" + strings.Repeat("0", 64)},
	})
	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestWebhook_GitHub_MissingHeaders(t *testing.T) {
	ts := testutil.NewTestServer(t)
	const secret = "test-webhook-secret"
	seedGitHubWebhookSecret(t, ts.DB, secret)

	body := []byte(`{"action":"opened"}`)
	sig := signGitHubPayload(secret, body)

	// Valid signature but no event/delivery headers → 400.
	rr := ts.DoWithHeaders(http.MethodPost, "/webhooks/github", body, http.Header{
		"X-Hub-Signature-256": []string{sig},
	})
	require.Equal(t, http.StatusBadRequest, rr.Code, "body=%s", rr.Body.String())
}
