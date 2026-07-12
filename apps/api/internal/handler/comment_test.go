package handler_test

import (
	"net/http"
	"testing"

	"github.com/Devlaner/devlane/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func commentBase(slug, projectID, issueID string) string {
	return "/api/workspaces/" + slug + "/projects/" + projectID + "/issues/" + issueID + "/comments/"
}

func TestComment_RequiresAuth(t *testing.T) {
	ts := testutil.NewTestServer(t)
	rr := ts.GET(commentBase("x", "00000000-0000-0000-0000-000000000000", "00000000-0000-0000-0000-000000000000"), "")
	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestComment_CRUD(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)
	issue := testutil.CreateIssue(t, ts.DB, w.Project.ID, w.Workspace.ID, w.User.ID)
	base := commentBase(w.Workspace.Slug, w.Project.ID.String(), issue.ID.String())

	// Create
	rr := ts.POST(base, map[string]any{"comment": "first comment"}, w.Session)
	require.Equal(t, http.StatusCreated, rr.Code, "body=%s", rr.Body.String())
	id, _ := testutil.MustJSONMap(t, rr)["id"].(string)
	require.NotEmpty(t, id)

	// List
	rr2 := ts.GET(base, w.Session)
	require.Equal(t, http.StatusOK, rr2.Code, "body=%s", rr2.Body.String())

	// Update
	rr3 := ts.PATCH(base+id+"/", map[string]any{"comment": "updated"}, w.Session)
	require.Equal(t, http.StatusOK, rr3.Code, "body=%s", rr3.Body.String())
	assert.Equal(t, "updated", testutil.MustJSONMap(t, rr3)["comment"])

	// Delete
	rr4 := ts.DELETE(base+id+"/", w.Session)
	require.Equal(t, http.StatusNoContent, rr4.Code, "body=%s", rr4.Body.String())
}

func TestComment_RecordsActivity(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)
	issue := testutil.CreateIssue(t, ts.DB, w.Project.ID, w.Workspace.ID, w.User.ID)
	base := commentBase(w.Workspace.Slug, w.Project.ID.String(), issue.ID.String())

	require.Equal(t, http.StatusCreated, ts.POST(base, map[string]any{"comment": "hi"}, w.Session).Code)

	acts := ts.GET(
		"/api/workspaces/"+w.Workspace.Slug+"/projects/"+w.Project.ID.String()+"/issues/"+issue.ID.String()+"/activities/",
		w.Session,
	)
	require.Equal(t, http.StatusOK, acts.Code, "body=%s", acts.Body.String())
	assert.Contains(t, acts.Body.String(), "comment_added")
}

func TestComment_Reactions(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)
	issue := testutil.CreateIssue(t, ts.DB, w.Project.ID, w.Workspace.ID, w.User.ID)
	comment := testutil.CreateComment(t, ts.DB, issue.ID, w.Project.ID, w.Workspace.ID, w.User.ID)
	base := commentBase(w.Workspace.Slug, w.Project.ID.String(), issue.ID.String()) + comment.ID.String() + "/reactions/"

	// List (empty)
	rr := ts.GET(base, w.Session)
	require.Equal(t, http.StatusOK, rr.Code, "body=%s", rr.Body.String())

	// Add reaction
	rr2 := ts.POST(base, map[string]any{"reaction": "👍"}, w.Session)
	require.Equal(t, http.StatusCreated, rr2.Code, "body=%s", rr2.Body.String())

	// Remove reaction
	rr3 := ts.DELETE(base+"%F0%9F%91%8D/", w.Session) // URL-encoded thumbs-up
	require.Equal(t, http.StatusNoContent, rr3.Code, "body=%s", rr3.Body.String())
}
