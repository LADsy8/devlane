package handler_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/Devlaner/devlane/api/internal/model"
	"github.com/Devlaner/devlane/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecentVisit_RequiresAuth(t *testing.T) {
	ts := testutil.NewTestServer(t)
	rr := ts.GET("/api/workspaces/x/recent-visits/", "")
	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestRecentVisit_ListEmpty(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)

	rr := ts.GET("/api/workspaces/"+w.Workspace.Slug+"/recent-visits/", w.Session)
	require.Equal(t, http.StatusOK, rr.Code, "body=%s", rr.Body.String())
}

func TestRecentVisit_RecordProject(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)

	rr := ts.POST("/api/workspaces/"+w.Workspace.Slug+"/recent-visits/", map[string]any{
		"entity_name":       "project",
		"entity_identifier": w.Project.ID.String(),
	}, w.Session)
	require.Equal(t, http.StatusNoContent, rr.Code, "body=%s", rr.Body.String())
}

func TestRecentVisit_RecordIssue(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)
	issue := testutil.CreateIssue(t, ts.DB, w.Project.ID, w.Workspace.ID, w.User.ID)

	rr := ts.POST("/api/workspaces/"+w.Workspace.Slug+"/recent-visits/", map[string]any{
		"entity_name":       "issue",
		"entity_identifier": issue.ID.String(),
		"project_id":        w.Project.ID.String(),
	}, w.Session)
	require.Equal(t, http.StatusNoContent, rr.Code, "body=%s", rr.Body.String())
}

func TestRecentVisit_RecordIssue_RejectsForeignWorkspace(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)

	otherOwner := testutil.CreateUser(t, ts.DB)
	otherWs := testutil.CreateWorkspace(t, ts.DB, otherOwner.ID)
	otherProject := testutil.CreateProject(t, ts.DB, otherWs.ID, otherOwner.ID)
	foreignIssue := testutil.CreateIssue(t, ts.DB, otherProject.ID, otherWs.ID, otherOwner.ID)

	rr := ts.POST("/api/workspaces/"+w.Workspace.Slug+"/recent-visits/", map[string]any{
		"entity_name":       "issue",
		"entity_identifier": foreignIssue.ID.String(),
	}, w.Session)
	require.Equal(t, http.StatusNotFound, rr.Code, "body=%s", rr.Body.String())
}

func TestRecentVisit_RecordProject_RejectsForeignWorkspace(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)

	otherOwner := testutil.CreateUser(t, ts.DB)
	otherWs := testutil.CreateWorkspace(t, ts.DB, otherOwner.ID)
	foreignProject := testutil.CreateProject(t, ts.DB, otherWs.ID, otherOwner.ID)

	rr := ts.POST("/api/workspaces/"+w.Workspace.Slug+"/recent-visits/", map[string]any{
		"entity_name":       "project",
		"entity_identifier": foreignProject.ID.String(),
	}, w.Session)
	require.Equal(t, http.StatusNotFound, rr.Code, "body=%s", rr.Body.String())
}

func TestRecentVisit_RecordPage_RejectsForeignWorkspace(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)

	otherOwner := testutil.CreateUser(t, ts.DB)
	otherWs := testutil.CreateWorkspace(t, ts.DB, otherOwner.ID)
	foreignPage := testutil.CreatePage(t, ts.DB, otherWs.ID, otherOwner.ID)

	rr := ts.POST("/api/workspaces/"+w.Workspace.Slug+"/recent-visits/", map[string]any{
		"entity_name":       "page",
		"entity_identifier": foreignPage.ID.String(),
	}, w.Session)
	require.Equal(t, http.StatusNotFound, rr.Code, "body=%s", rr.Body.String())
}

func TestRecentVisit_RecordIssue_RejectsForeignProjectID(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)
	issue := testutil.CreateIssue(t, ts.DB, w.Project.ID, w.Workspace.ID, w.User.ID)

	otherOwner := testutil.CreateUser(t, ts.DB)
	otherWs := testutil.CreateWorkspace(t, ts.DB, otherOwner.ID)
	foreignProject := testutil.CreateProject(t, ts.DB, otherWs.ID, otherOwner.ID)

	rr := ts.POST("/api/workspaces/"+w.Workspace.Slug+"/recent-visits/", map[string]any{
		"entity_name":       "issue",
		"entity_identifier": issue.ID.String(),
		"project_id":        foreignProject.ID.String(),
	}, w.Session)
	require.Equal(t, http.StatusNotFound, rr.Code, "body=%s", rr.Body.String())
}

// TestRecentVisit_List_DoesNotLeakPreExistingForeignRow proves the read-side
// defense in depth: even a recent-visit row that already points at a
// foreign-workspace entity (as could exist from before this fix, or from a
// bypass of RecordVisit) must not have its title/identifier reflected back.
func TestRecentVisit_List_DoesNotLeakPreExistingForeignRow(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)

	otherOwner := testutil.CreateUser(t, ts.DB)
	otherWs := testutil.CreateWorkspace(t, ts.DB, otherOwner.ID)
	otherProject := testutil.CreateProject(t, ts.DB, otherWs.ID, otherOwner.ID)
	foreignIssue := testutil.CreateIssue(t, ts.DB, otherProject.ID, otherWs.ID, otherOwner.ID)

	// Insert the bad row directly, bypassing RecordVisit's new validation, to
	// simulate data written before this fix shipped.
	badVisit := &model.UserRecentVisit{
		WorkspaceID:      w.Workspace.ID,
		UserID:           w.User.ID,
		EntityName:       "issue",
		EntityIdentifier: &foreignIssue.ID,
		LastVisitedAt:    time.Now(),
	}
	require.NoError(t, ts.DB.WithContext(context.Background()).Create(badVisit).Error)

	rr := ts.GET("/api/workspaces/"+w.Workspace.Slug+"/recent-visits/", w.Session)
	require.Equal(t, http.StatusOK, rr.Code, "body=%s", rr.Body.String())
	list := testutil.DecodeJSON[[]map[string]any](t, rr)
	require.Len(t, list, 1)
	assert.Empty(t, list[0]["display_title"])
	assert.Empty(t, list[0]["display_identifier"])
}
