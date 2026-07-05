package handler_test

import (
	"net/http"
	"testing"

	"github.com/Devlaner/devlane/api/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestIssue_CreateRejectsForeignRelations(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)

	// A second project in the same workspace, and one in another workspace.
	otherProject := testutil.CreateProject(t, ts.DB, w.Workspace.ID, w.User.ID)
	foreignState := testutil.CreateState(t, ts.DB, otherProject.ID, w.Workspace.ID)
	foreignLabel := testutil.CreateLabel(t, ts.DB, otherProject.ID, w.Workspace.ID)
	foreignParent := testutil.CreateIssue(t, ts.DB, otherProject.ID, w.Workspace.ID, w.User.ID)
	nonMember := testutil.CreateUser(t, ts.DB)

	base := "/api/workspaces/" + w.Workspace.Slug + "/projects/" + w.Project.ID.String() + "/issues/"

	require.Equal(t, http.StatusBadRequest,
		ts.POST(base, map[string]any{"name": "x", "state_id": foreignState.ID.String()}, w.Session).Code,
		"a state from another project must be rejected")
	require.Equal(t, http.StatusBadRequest,
		ts.POST(base, map[string]any{"name": "x", "label_ids": []string{foreignLabel.ID.String()}}, w.Session).Code,
		"a label from another project must be rejected")
	require.Equal(t, http.StatusBadRequest,
		ts.POST(base, map[string]any{"name": "x", "parent_id": foreignParent.ID.String()}, w.Session).Code,
		"a parent from another project must be rejected")
	require.Equal(t, http.StatusBadRequest,
		ts.POST(base, map[string]any{"name": "x", "assignee_ids": []string{nonMember.ID.String()}}, w.Session).Code,
		"an assignee who isn't a workspace member must be rejected")

	// A well-scoped create still works.
	okState := testutil.CreateState(t, ts.DB, w.Project.ID, w.Workspace.ID)
	okLabel := testutil.CreateLabel(t, ts.DB, w.Project.ID, w.Workspace.ID)
	okParent := testutil.CreateIssue(t, ts.DB, w.Project.ID, w.Workspace.ID, w.User.ID)
	rr := ts.POST(base, map[string]any{
		"name":         "valid",
		"state_id":     okState.ID.String(),
		"label_ids":    []string{okLabel.ID.String()},
		"assignee_ids": []string{w.User.ID.String()},
		"parent_id":    okParent.ID.String(),
	}, w.Session)
	require.Equal(t, http.StatusCreated, rr.Code, "body=%s", rr.Body.String())
}

func TestIssue_UpdateRejectsForeignRelations(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)
	otherProject := testutil.CreateProject(t, ts.DB, w.Workspace.ID, w.User.ID)
	foreignState := testutil.CreateState(t, ts.DB, otherProject.ID, w.Workspace.ID)
	foreignLabel := testutil.CreateLabel(t, ts.DB, otherProject.ID, w.Workspace.ID)
	nonMember := testutil.CreateUser(t, ts.DB)

	issue := testutil.CreateIssue(t, ts.DB, w.Project.ID, w.Workspace.ID, w.User.ID)
	base := "/api/workspaces/" + w.Workspace.Slug + "/projects/" + w.Project.ID.String() +
		"/issues/" + issue.ID.String() + "/"

	require.Equal(t, http.StatusBadRequest,
		ts.PATCH(base, map[string]any{"state_id": foreignState.ID.String()}, w.Session).Code)
	require.Equal(t, http.StatusBadRequest,
		ts.PATCH(base, map[string]any{"label_ids": []string{foreignLabel.ID.String()}}, w.Session).Code)
	require.Equal(t, http.StatusBadRequest,
		ts.PATCH(base, map[string]any{"assignee_ids": []string{nonMember.ID.String()}}, w.Session).Code)

	// A parent from another project is rejected.
	foreignParent := testutil.CreateIssue(t, ts.DB, otherProject.ID, w.Workspace.ID, w.User.ID)
	require.Equal(t, http.StatusBadRequest,
		ts.PATCH(base, map[string]any{"parent_id": foreignParent.ID.String()}, w.Session).Code)

	// An issue can't be made its own parent.
	require.Equal(t, http.StatusBadRequest,
		ts.PATCH(base, map[string]any{"parent_id": issue.ID.String()}, w.Session).Code)

	// A well-scoped update still works.
	okState := testutil.CreateState(t, ts.DB, w.Project.ID, w.Workspace.ID)
	require.Equal(t, http.StatusOK,
		ts.PATCH(base, map[string]any{"state_id": okState.ID.String()}, w.Session).Code)
}
