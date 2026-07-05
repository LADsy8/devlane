package handler_test

import (
	"net/http"
	"testing"

	"github.com/Devlaner/devlane/api/internal/model"
	"github.com/Devlaner/devlane/api/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// A regular workspace member who creates a project should become a project
// admin, so they can immediately manage the project they just made.
func TestProject_CreatorBecomesProjectAdmin(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)

	member := testutil.CreateUser(t, ts.DB)
	testutil.AddWorkspaceMember(t, ts.DB, w.Workspace.ID, member.ID, testutil.RoleMember)
	session := testutil.LoginAs(t, ts.DB, member)

	base := "/api/workspaces/" + w.Workspace.Slug + "/projects/"
	rr := ts.POST(base, map[string]any{"name": "Member Project", "identifier": "MBR"}, session)
	require.Equal(t, http.StatusCreated, rr.Code, "body=%s", rr.Body.String())
	projectID, _ := testutil.MustJSONMap(t, rr)["id"].(string)
	require.NotEmpty(t, projectID)

	// A project_members row exists for the creator with at least admin role.
	var pm model.ProjectMember
	require.NoError(t, ts.DB.
		Where("project_id = ? AND member_id = ? AND deleted_at IS NULL", projectID, member.ID).
		First(&pm).Error)
	require.GreaterOrEqual(t, pm.Role, model.RoleAdmin)

	// The creator can immediately save project settings (admin-only action).
	rr2 := ts.PATCH(base+projectID+"/", map[string]any{"name": "Renamed"}, session)
	require.Equal(t, http.StatusOK, rr2.Code, "body=%s", rr2.Body.String())
}

// Guard: creating a project must not add a stray membership for anyone else.
func TestProject_CreateAddsOnlyCreator(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)

	base := "/api/workspaces/" + w.Workspace.Slug + "/projects/"
	rr := ts.POST(base, map[string]any{"name": "Solo", "identifier": "SOLO"}, w.Session)
	require.Equal(t, http.StatusCreated, rr.Code, "body=%s", rr.Body.String())
	projectID, _ := testutil.MustJSONMap(t, rr)["id"].(string)

	var count int64
	require.NoError(t, ts.DB.Model(&model.ProjectMember{}).
		Where("project_id = ? AND deleted_at IS NULL", projectID).Count(&count).Error)
	require.Equal(t, int64(1), count)

	var pm model.ProjectMember
	require.NoError(t, ts.DB.
		Where("project_id = ? AND deleted_at IS NULL", projectID).First(&pm).Error)
	require.NotNil(t, pm.MemberID)
	require.Equal(t, w.User.ID, *pm.MemberID)
	// Sanity: the created uuid parses.
	require.NotEqual(t, uuid.Nil, uuid.MustParse(projectID))
}
