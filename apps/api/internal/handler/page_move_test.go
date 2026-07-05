package handler_test

import (
	"net/http"
	"testing"

	"github.com/Devlaner/devlane/api/internal/model"
	"github.com/Devlaner/devlane/api/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func linkPageToProject(t *testing.T, ts *testutil.TestServer, projectID, pageID, workspaceID uuid.UUID) {
	t.Helper()
	require.NoError(t, ts.DB.Create(&model.ProjectPage{
		ProjectID:   projectID,
		PageID:      pageID,
		WorkspaceID: workspaceID,
	}).Error)
}

func countPageLinks(t *testing.T, ts *testutil.TestServer, projectID uuid.UUID, pageIDs ...uuid.UUID) int64 {
	t.Helper()
	var n int64
	require.NoError(t, ts.DB.Model(&model.ProjectPage{}).
		Where("project_id = ? AND page_id IN ?", projectID, pageIDs).
		Count(&n).Error)
	return n
}

func TestPage_MoveToProject(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)
	target := testutil.CreateProject(t, ts.DB, w.Workspace.ID, w.User.ID)

	// A small tree in the source project: topParent > page > child. Only `page`
	// and its subtree move; topParent stays behind.
	topParent := testutil.CreatePage(t, ts.DB, w.Workspace.ID, w.User.ID)
	linkPageToProject(t, ts, w.Project.ID, topParent.ID, w.Workspace.ID)

	page := testutil.CreatePage(t, ts.DB, w.Workspace.ID, w.User.ID)
	require.NoError(t, ts.DB.Model(page).Update("parent_id", topParent.ID).Error)
	linkPageToProject(t, ts, w.Project.ID, page.ID, w.Workspace.ID)

	child := testutil.CreatePage(t, ts.DB, w.Workspace.ID, w.User.ID)
	require.NoError(t, ts.DB.Model(child).Update("parent_id", page.ID).Error)
	linkPageToProject(t, ts, w.Project.ID, child.ID, w.Workspace.ID)

	base := "/api/workspaces/" + w.Workspace.Slug + "/pages/" + page.ID.String()
	rr := ts.POST(base+"/move/", map[string]any{"target_project_id": target.ID.String()}, w.Session)
	require.Equal(t, http.StatusOK, rr.Code, "body=%s", rr.Body.String())

	// page + child are now linked to the target project and unlinked from source.
	require.Equal(t, int64(2), countPageLinks(t, ts, target.ID, page.ID, child.ID))
	require.Zero(t, countPageLinks(t, ts, w.Project.ID, page.ID, child.ID))

	// The moved root is detached from its old-project parent; the child keeps its
	// parent (which moved along with it).
	var movedPage, movedChild model.Page
	require.NoError(t, ts.DB.First(&movedPage, "id = ?", page.ID).Error)
	require.NoError(t, ts.DB.First(&movedChild, "id = ?", child.ID).Error)
	require.Nil(t, movedPage.ParentID, "moved root should be detached from its source-project parent")
	require.NotNil(t, movedChild.ParentID)
	require.Equal(t, page.ID, *movedChild.ParentID)

	// topParent is untouched: still in the source project.
	require.Equal(t, int64(1), countPageLinks(t, ts, w.Project.ID, topParent.ID))
}

func TestPage_MoveRequiresTarget(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)
	page := testutil.CreatePage(t, ts.DB, w.Workspace.ID, w.User.ID)
	linkPageToProject(t, ts, w.Project.ID, page.ID, w.Workspace.ID)
	base := "/api/workspaces/" + w.Workspace.Slug + "/pages/" + page.ID.String()
	rr := ts.POST(base+"/move/", map[string]any{}, w.Session)
	require.Equal(t, http.StatusBadRequest, rr.Code, "body=%s", rr.Body.String())
}

func TestPage_MoveToSameProjectRejected(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)
	page := testutil.CreatePage(t, ts.DB, w.Workspace.ID, w.User.ID)
	linkPageToProject(t, ts, w.Project.ID, page.ID, w.Workspace.ID)
	base := "/api/workspaces/" + w.Workspace.Slug + "/pages/" + page.ID.String()
	rr := ts.POST(base+"/move/", map[string]any{"target_project_id": w.Project.ID.String()}, w.Session)
	require.Equal(t, http.StatusBadRequest, rr.Code, "body=%s", rr.Body.String())
}

func TestPage_MoveCrossWorkspaceRejected(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)
	otherUser := testutil.CreateUser(t, ts.DB)
	otherWS := testutil.CreateWorkspace(t, ts.DB, otherUser.ID)
	foreign := testutil.CreateProject(t, ts.DB, otherWS.ID, otherUser.ID)

	page := testutil.CreatePage(t, ts.DB, w.Workspace.ID, w.User.ID)
	linkPageToProject(t, ts, w.Project.ID, page.ID, w.Workspace.ID)
	base := "/api/workspaces/" + w.Workspace.Slug + "/pages/" + page.ID.String()
	rr := ts.POST(base+"/move/", map[string]any{"target_project_id": foreign.ID.String()}, w.Session)
	require.Equal(t, http.StatusNotFound, rr.Code, "body=%s", rr.Body.String())
}
