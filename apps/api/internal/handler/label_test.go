package handler_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/Devlaner/devlane/api/internal/model"
	"github.com/Devlaner/devlane/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLabel_RequiresAuth(t *testing.T) {
	ts := testutil.NewTestServer(t)
	rr := ts.GET("/api/workspaces/x/projects/00000000-0000-0000-0000-000000000000/issue-labels/", "")
	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestLabel_CRUD(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)
	base := "/api/workspaces/" + w.Workspace.Slug + "/projects/" + w.Project.ID.String() + "/issue-labels/"

	// Create
	rr := ts.POST(base, map[string]any{"name": "bug", "color": "#ff0000"}, w.Session)
	require.Equal(t, http.StatusCreated, rr.Code, "body=%s", rr.Body.String())
	id, _ := testutil.MustJSONMap(t, rr)["id"].(string)
	require.NotEmpty(t, id)

	// List
	rr2 := ts.GET(base, w.Session)
	require.Equal(t, http.StatusOK, rr2.Code)
	assert.Len(t, testutil.DecodeJSON[[]map[string]any](t, rr2), 1)

	// Update
	rr3 := ts.PATCH(base+id+"/", map[string]any{"name": "feature"}, w.Session)
	require.Equal(t, http.StatusOK, rr3.Code)
	assert.Equal(t, "feature", testutil.MustJSONMap(t, rr3)["name"])

	// Delete
	rr4 := ts.DELETE(base+id+"/", w.Session)
	require.Equal(t, http.StatusNoContent, rr4.Code)
}

// TestLabel_WorkspaceLevel_RejectsForeignWorkspace proves a workspace-level
// label (ProjectID == nil) from a foreign workspace can't be read, updated,
// or deleted just by supplying its UUID alongside any project in a workspace
// the caller does belong to.
func TestLabel_WorkspaceLevel_RejectsForeignWorkspace(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)
	base := "/api/workspaces/" + w.Workspace.Slug + "/projects/" + w.Project.ID.String() + "/issue-labels/"

	otherOwner := testutil.CreateUser(t, ts.DB)
	otherWs := testutil.CreateWorkspace(t, ts.DB, otherOwner.ID)
	foreignLabel := &model.Label{
		Name:        "foreign workspace label",
		Color:       "#00ff00",
		ProjectID:   nil,
		WorkspaceID: otherWs.ID,
	}
	require.NoError(t, ts.DB.WithContext(context.Background()).Create(foreignLabel).Error)

	rr := ts.PATCH(base+foreignLabel.ID.String()+"/", map[string]any{"name": "hijacked"}, w.Session)
	require.Equal(t, http.StatusNotFound, rr.Code, "body=%s", rr.Body.String())

	rr2 := ts.DELETE(base+foreignLabel.ID.String()+"/", w.Session)
	require.Equal(t, http.StatusNotFound, rr2.Code, "body=%s", rr2.Body.String())
}
