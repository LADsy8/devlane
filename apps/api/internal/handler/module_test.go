package handler_test

import (
	"net/http"
	"testing"

	"github.com/Devlaner/devlane/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func moduleBase(slug, projectID string) string {
	return "/api/workspaces/" + slug + "/projects/" + projectID + "/modules/"
}

func TestModule_RequiresAuth(t *testing.T) {
	ts := testutil.NewTestServer(t)
	rr := ts.GET(moduleBase("x", "00000000-0000-0000-0000-000000000000"), "")
	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestModule_CRUD(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)
	base := moduleBase(w.Workspace.Slug, w.Project.ID.String())

	// Create
	rr := ts.POST(base, map[string]any{
		"name":        "Auth Module",
		"description": "All auth-related work",
	}, w.Session)
	require.Equal(t, http.StatusCreated, rr.Code, "body=%s", rr.Body.String())
	id, _ := testutil.MustJSONMap(t, rr)["id"].(string)
	require.NotEmpty(t, id)

	// List
	rr2 := ts.GET(base, w.Session)
	require.Equal(t, http.StatusOK, rr2.Code)

	// Get
	rr3 := ts.GET(base+id+"/", w.Session)
	require.Equal(t, http.StatusOK, rr3.Code)
	assert.Equal(t, "Auth Module", testutil.MustJSONMap(t, rr3)["name"])

	// Update
	rr4 := ts.PATCH(base+id+"/", map[string]any{"name": "Auth Module v2"}, w.Session)
	require.Equal(t, http.StatusOK, rr4.Code)
	assert.Equal(t, "Auth Module v2", testutil.MustJSONMap(t, rr4)["name"])

	// Delete
	rr5 := ts.DELETE(base+id+"/", w.Session)
	require.Equal(t, http.StatusNoContent, rr5.Code)
}

func TestModule_RejectsInvalidDates(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)
	base := moduleBase(w.Workspace.Slug, w.Project.ID.String())

	// Target before start is rejected.
	rr := ts.POST(base, map[string]any{
		"name":        "Backwards",
		"start_date":  "2026-02-10",
		"target_date": "2026-02-01",
	}, w.Session)
	require.Equal(t, http.StatusBadRequest, rr.Code, "body=%s", rr.Body.String())

	// A valid range is accepted.
	ok := ts.POST(base, map[string]any{
		"name":        "Module",
		"start_date":  "2026-02-01",
		"target_date": "2026-02-10",
	}, w.Session)
	require.Equal(t, http.StatusCreated, ok.Code, "body=%s", ok.Body.String())
}

func TestModule_Issues_AddListRemove(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)
	mod := testutil.CreateModule(t, ts.DB, w.Project.ID, w.Workspace.ID)
	issue := testutil.CreateIssue(t, ts.DB, w.Project.ID, w.Workspace.ID, w.User.ID)
	base := moduleBase(w.Workspace.Slug, w.Project.ID.String()) + mod.ID.String() + "/issues/"

	// Add
	rr := ts.POST(base, map[string]any{"issue_id": issue.ID.String()}, w.Session)
	require.Truef(t, rr.Code < 400, "unexpected status %d body=%s", rr.Code, rr.Body.String())

	// List
	rr2 := ts.GET(base, w.Session)
	require.Equal(t, http.StatusOK, rr2.Code, "body=%s", rr2.Body.String())

	// Remove
	rr3 := ts.DELETE(base+issue.ID.String()+"/", w.Session)
	require.Equal(t, http.StatusNoContent, rr3.Code, "body=%s", rr3.Body.String())
}
