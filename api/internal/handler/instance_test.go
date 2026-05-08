package handler_test

import (
	"net/http"
	"testing"

	"github.com/Devlaner/devlane/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstance_SetupStatus_FreshDB(t *testing.T) {
	ts := testutil.NewTestServer(t)

	rr := ts.GET("/api/instance/setup-status/", "")
	require.Equal(t, http.StatusOK, rr.Code, "body=%s", rr.Body.String())
	body := testutil.MustJSONMap(t, rr)
	assert.Equal(t, true, body["setup_required"])
}

func TestInstance_SetupStatus_AfterSetup(t *testing.T) {
	ts := testutil.NewTestServer(t)
	// Seeding a user marks the instance as set up.
	testutil.CreateUser(t, ts.DB)

	rr := ts.GET("/api/instance/setup-status/", "")
	require.Equal(t, http.StatusOK, rr.Code)
	body := testutil.MustJSONMap(t, rr)
	assert.Equal(t, false, body["setup_required"])
}

func TestInstance_Setup_Success(t *testing.T) {
	ts := testutil.NewTestServer(t)

	rr := ts.POST("/api/instance/setup/", map[string]any{
		"first_name":   "Ada",
		"last_name":    "Lovelace",
		"email":        "ada@test.local",
		"password":     "S3cur3!Pass",
		"company_name": "Analytical Engine Co.",
	}, "")
	require.Equal(t, http.StatusCreated, rr.Code, "body=%s", rr.Body.String())

	user := testutil.MustJSONMap(t, rr)
	assert.Equal(t, "ada@test.local", user["email"])
	assert.NotEmpty(t, user["id"])

	// Setup-status should now report no setup needed.
	rr2 := ts.GET("/api/instance/setup-status/", "")
	require.Equal(t, http.StatusOK, rr2.Code)
	assert.Equal(t, false, testutil.MustJSONMap(t, rr2)["setup_required"])
}

func TestInstance_Setup_RejectsSecondCall(t *testing.T) {
	ts := testutil.NewTestServer(t)
	testutil.CreateUser(t, ts.DB)

	rr := ts.POST("/api/instance/setup/", map[string]any{
		"first_name": "Ada",
		"last_name":  "Lovelace",
		"email":      "ada2@test.local",
		"password":   "S3cur3!Pass",
	}, "")
	require.Equal(t, http.StatusForbidden, rr.Code)
}

func TestInstance_Setup_RejectsWeakPassword(t *testing.T) {
	ts := testutil.NewTestServer(t)

	rr := ts.POST("/api/instance/setup/", map[string]any{
		"first_name": "Ada",
		"last_name":  "Lovelace",
		"email":      "ada@test.local",
		"password":   "tooweak",
	}, "")
	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestInstance_Settings_RequiresAuth(t *testing.T) {
	ts := testutil.NewTestServer(t)

	rr := ts.GET("/api/instance/settings/", "")
	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestInstance_Settings_GetWithAuth(t *testing.T) {
	ts := testutil.NewTestServer(t)
	user := testutil.CreateUser(t, ts.DB)
	session := testutil.LoginAs(t, ts.DB, user)

	rr := ts.GET("/api/instance/settings/", session)
	require.Equal(t, http.StatusOK, rr.Code, "body=%s", rr.Body.String())

	body := testutil.MustJSONMap(t, rr)
	// Defaults are filled in for every section even on a fresh DB.
	for _, key := range []string{"general", "email", "auth", "oauth", "ai", "image", "github_app"} {
		assert.NotNil(t, body[key], "missing settings section %q", key)
	}
}

func TestInstance_Settings_UpdateGeneral(t *testing.T) {
	ts := testutil.NewTestServer(t)
	user := testutil.CreateUser(t, ts.DB)
	session := testutil.LoginAs(t, ts.DB, user)

	rr := ts.PATCH("/api/instance/settings/general", map[string]any{
		"value": map[string]any{
			"instance_name":                   "Acme Inc.",
			"only_admin_can_create_workspace": true,
		},
	}, session)
	require.Equal(t, http.StatusOK, rr.Code, "body=%s", rr.Body.String())

	out := testutil.MustJSONMap(t, rr)
	assert.Equal(t, "general", out["key"])
	val, _ := out["value"].(map[string]any)
	require.NotNil(t, val)
	assert.Equal(t, "Acme Inc.", val["instance_name"])
	assert.Equal(t, true, val["only_admin_can_create_workspace"])
}

func TestInstance_Settings_UpdateInvalidKey(t *testing.T) {
	ts := testutil.NewTestServer(t)
	user := testutil.CreateUser(t, ts.DB)
	session := testutil.LoginAs(t, ts.DB, user)

	rr := ts.PATCH("/api/instance/settings/not-a-real-section", map[string]any{
		"value": map[string]any{},
	}, session)
	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestInstance_Unsplash_NotConfigured(t *testing.T) {
	ts := testutil.NewTestServer(t)
	user := testutil.CreateUser(t, ts.DB)
	session := testutil.LoginAs(t, ts.DB, user)

	rr := ts.GET("/api/instance/unsplash/search?q=cats", session)
	require.Equal(t, http.StatusBadRequest, rr.Code, "body=%s", rr.Body.String())
}
