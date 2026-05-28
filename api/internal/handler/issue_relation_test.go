package handler_test

import (
	"net/http"
	"testing"

	"github.com/Devlaner/devlane/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func relationBase(slug, projectID, issueID string) string {
	return issueBase(slug, projectID) + issueID + "/issue-relation/"
}

func removeRelationURL(slug, projectID, issueID string) string {
	return issueBase(slug, projectID) + issueID + "/remove-relation/"
}

func TestIssueRelation_RequiresAuth(t *testing.T) {
	ts := testutil.NewTestServer(t)
	nilID := "00000000-0000-0000-0000-000000000000"
	rr := ts.GET(relationBase("x", nilID, nilID), "")
	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestIssueRelation_ListEmpty(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)
	issue := testutil.CreateIssue(t, ts.DB, w.Project.ID, w.Workspace.ID, w.User.ID)

	rr := ts.GET(relationBase(w.Workspace.Slug, w.Project.ID.String(), issue.ID.String()), w.Session)
	require.Equal(t, http.StatusOK, rr.Code, "body=%s", rr.Body.String())

	body := testutil.MustJSONMap(t, rr)
	for _, key := range []string{"blocking", "blocked_by", "duplicate", "relates_to"} {
		val, ok := body[key]
		require.True(t, ok, "missing key %s", key)
		arr, ok := val.([]interface{})
		require.True(t, ok, "key %s is not an array", key)
		assert.Empty(t, arr, "key %s should be empty", key)
	}
}

func TestIssueRelation_CreateAndList(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)
	issueA := testutil.CreateIssue(t, ts.DB, w.Project.ID, w.Workspace.ID, w.User.ID)
	issueB := testutil.CreateIssue(t, ts.DB, w.Project.ID, w.Workspace.ID, w.User.ID)

	base := relationBase(w.Workspace.Slug, w.Project.ID.String(), issueA.ID.String())

	// Create: A blocks B
	rr := ts.POST(base, map[string]any{
		"relation_type": "blocking",
		"issues":        []string{issueB.ID.String()},
	}, w.Session)
	require.Equal(t, http.StatusCreated, rr.Code, "body=%s", rr.Body.String())

	// List A's relations — should show B under "blocking"
	rr2 := ts.GET(base, w.Session)
	require.Equal(t, http.StatusOK, rr2.Code, "body=%s", rr2.Body.String())
	body := testutil.MustJSONMap(t, rr2)
	blocking, _ := body["blocking"].([]interface{})
	require.Len(t, blocking, 1)

	// List B's relations — should show A under "blocked_by" (reverse was stored)
	baseB := relationBase(w.Workspace.Slug, w.Project.ID.String(), issueB.ID.String())
	rr3 := ts.GET(baseB, w.Session)
	require.Equal(t, http.StatusOK, rr3.Code, "body=%s", rr3.Body.String())
	bodyB := testutil.MustJSONMap(t, rr3)
	blockedBy, _ := bodyB["blocked_by"].([]interface{})
	require.Len(t, blockedBy, 1)
}

func TestIssueRelation_CreateDuplicate_IsIdempotent(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)
	issueA := testutil.CreateIssue(t, ts.DB, w.Project.ID, w.Workspace.ID, w.User.ID)
	issueB := testutil.CreateIssue(t, ts.DB, w.Project.ID, w.Workspace.ID, w.User.ID)

	base := relationBase(w.Workspace.Slug, w.Project.ID.String(), issueA.ID.String())
	body := map[string]any{"relation_type": "relates_to", "issues": []string{issueB.ID.String()}}

	ts.POST(base, body, w.Session)       // first insert
	rr := ts.POST(base, body, w.Session) // duplicate
	require.Equal(t, http.StatusCreated, rr.Code, "duplicate insert should not error")

	// Still only one relation
	rr2 := ts.GET(base, w.Session)
	require.Equal(t, http.StatusOK, rr2.Code)
	got := testutil.MustJSONMap(t, rr2)
	relatesTo, _ := got["relates_to"].([]interface{})
	assert.Len(t, relatesTo, 1, "duplicate insert should not create a second row")
}

func TestIssueRelation_Remove(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)
	issueA := testutil.CreateIssue(t, ts.DB, w.Project.ID, w.Workspace.ID, w.User.ID)
	issueB := testutil.CreateIssue(t, ts.DB, w.Project.ID, w.Workspace.ID, w.User.ID)

	createBase := relationBase(w.Workspace.Slug, w.Project.ID.String(), issueA.ID.String())
	removeURL := removeRelationURL(w.Workspace.Slug, w.Project.ID.String(), issueA.ID.String())

	// Create the relation first
	rr := ts.POST(createBase, map[string]any{
		"relation_type": "duplicate",
		"issues":        []string{issueB.ID.String()},
	}, w.Session)
	require.Equal(t, http.StatusCreated, rr.Code)

	// Remove it
	rr2 := ts.POST(removeURL, map[string]any{
		"relation_type": "duplicate",
		"related_issue": issueB.ID.String(),
	}, w.Session)
	require.Equal(t, http.StatusNoContent, rr2.Code, "body=%s", rr2.Body.String())

	// List should be empty again
	rr3 := ts.GET(createBase, w.Session)
	require.Equal(t, http.StatusOK, rr3.Code)
	body := testutil.MustJSONMap(t, rr3)
	duplicate, _ := body["duplicate"].([]interface{})
	assert.Empty(t, duplicate)
}

func TestIssueRelation_NonMember404(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)
	issue := testutil.CreateIssue(t, ts.DB, w.Project.ID, w.Workspace.ID, w.User.ID)
	stranger := testutil.CreateUser(t, ts.DB)
	strangerSession := testutil.LoginAs(t, ts.DB, stranger)

	rr := ts.GET(relationBase(w.Workspace.Slug, w.Project.ID.String(), issue.ID.String()), strangerSession)
	require.Equal(t, http.StatusNotFound, rr.Code)
}
