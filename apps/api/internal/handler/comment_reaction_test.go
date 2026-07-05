package handler_test

import (
	"net/http"
	"testing"

	"github.com/Devlaner/devlane/api/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestComment_AddReaction_StatusCodes(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)
	issue := testutil.CreateIssue(t, ts.DB, w.Project.ID, w.Workspace.ID, w.User.ID)
	comment := testutil.CreateComment(t, ts.DB, issue.ID, w.Project.ID, w.Workspace.ID, w.User.ID)

	base := "/api/workspaces/" + w.Workspace.Slug + "/projects/" + w.Project.ID.String() +
		"/issues/" + issue.ID.String() + "/comments/" + comment.ID.String() + "/reactions/"

	// First reaction succeeds.
	rr := ts.POST(base, map[string]any{"reaction": "👍"}, w.Session)
	require.Equal(t, http.StatusCreated, rr.Code, "body=%s", rr.Body.String())

	// Reacting again with the same emoji is a conflict, not a leaked 409-for-all.
	rr2 := ts.POST(base, map[string]any{"reaction": "👍"}, w.Session)
	require.Equal(t, http.StatusConflict, rr2.Code, "body=%s", rr2.Body.String())

	// A missing comment is a 404, not a 409.
	missing := "/api/workspaces/" + w.Workspace.Slug + "/projects/" + w.Project.ID.String() +
		"/issues/" + issue.ID.String() + "/comments/" + uuid.NewString() + "/reactions/"
	rr3 := ts.POST(missing, map[string]any{"reaction": "👍"}, w.Session)
	require.Equal(t, http.StatusNotFound, rr3.Code, "body=%s", rr3.Body.String())
}

func TestComment_RemoveReaction_StatusCodes(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)
	issue := testutil.CreateIssue(t, ts.DB, w.Project.ID, w.Workspace.ID, w.User.ID)
	comment := testutil.CreateComment(t, ts.DB, issue.ID, w.Project.ID, w.Workspace.ID, w.User.ID)

	base := "/api/workspaces/" + w.Workspace.Slug + "/projects/" + w.Project.ID.String() +
		"/issues/" + issue.ID.String() + "/comments/" + comment.ID.String() + "/reactions/"

	// Add then remove succeeds.
	require.Equal(t, http.StatusCreated, ts.POST(base, map[string]any{"reaction": "🎉"}, w.Session).Code)
	rr := ts.DELETE(base+"%F0%9F%8E%89/", w.Session)
	require.Equal(t, http.StatusNoContent, rr.Code, "body=%s", rr.Body.String())

	// Removing from a missing comment is a 404, not a generic 500.
	missing := "/api/workspaces/" + w.Workspace.Slug + "/projects/" + w.Project.ID.String() +
		"/issues/" + issue.ID.String() + "/comments/" + uuid.NewString() + "/reactions/x/"
	rr2 := ts.DELETE(missing, w.Session)
	require.Equal(t, http.StatusNotFound, rr2.Code, "body=%s", rr2.Body.String())
}
