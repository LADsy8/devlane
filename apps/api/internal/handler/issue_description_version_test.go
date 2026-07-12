package handler_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Devlaner/devlane/api/internal/model"
	"github.com/Devlaner/devlane/api/internal/testutil"
	"github.com/stretchr/testify/require"
)

// Editing a work item's description records a version of the previous text, the
// history lists them newest-first, and restoring an older version reverts the
// description (snapshotting the current one first so the restore is undoable).
func TestIssue_DescriptionVersionHistoryAndRestore(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)
	issue := testutil.CreateIssue(t, ts.DB, w.Project.ID, w.Workspace.ID, w.User.ID)

	base := "/api/workspaces/" + w.Workspace.Slug + "/projects/" + w.Project.ID.String() +
		"/issues/" + issue.ID.String() + "/"
	versionsURL := base + "description-versions/"

	// First edit: snapshots the empty original. Second edit: snapshots "v1".
	require.Equal(t, http.StatusOK, ts.PATCH(base, map[string]any{"description": "<p>v1</p>"}, w.Session).Code)
	require.Equal(t, http.StatusOK, ts.PATCH(base, map[string]any{"description": "<p>v2</p>"}, w.Session).Code)

	var versions []model.IssueDescriptionVersion
	rr := ts.GET(versionsURL, w.Session)
	require.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &versions))
	require.Len(t, versions, 2, "two edits should leave two prior-description snapshots")
	require.Equal(t, "<p>v1</p>", versions[0].DescriptionHTML, "newest snapshot is the pre-v2 text")

	// Restore the v1 snapshot; the description reverts and the current v2 is kept.
	restoreURL := versionsURL + versions[0].ID.String() + "/restore/"
	require.Equal(t, http.StatusOK, ts.POST(restoreURL, nil, w.Session).Code)

	var got model.Issue
	require.NoError(t, ts.DB.First(&got, "id = ?", issue.ID).Error)
	require.Equal(t, "<p>v1</p>", got.DescriptionHTML, "restore should revert the description")

	// The pre-restore v2 was snapshotted, so history now has three entries.
	rr2 := ts.GET(versionsURL, w.Session)
	var versions2 []model.IssueDescriptionVersion
	require.NoError(t, json.Unmarshal(rr2.Body.Bytes(), &versions2))
	require.Len(t, versions2, 3)
	require.Equal(t, "<p>v2</p>", versions2[0].DescriptionHTML, "restore snapshots the replaced text")
}
