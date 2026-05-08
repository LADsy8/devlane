package handler_test

import (
	"net/http"
	"testing"

	"github.com/Devlaner/devlane/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspace_List_RequiresAuth(t *testing.T) {
	ts := testutil.NewTestServer(t)
	rr := ts.GET("/api/users/me/workspaces/", "")
	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestWorkspace_List_OnlyOwn(t *testing.T) {
	ts := testutil.NewTestServer(t)
	owner := testutil.CreateUser(t, ts.DB)
	w := testutil.CreateWorkspace(t, ts.DB, owner.ID)

	other := testutil.CreateUser(t, ts.DB)
	_ = testutil.CreateWorkspace(t, ts.DB, other.ID) // unrelated workspace

	session := testutil.LoginAs(t, ts.DB, owner)
	rr := ts.GET("/api/users/me/workspaces/", session)
	require.Equal(t, http.StatusOK, rr.Code, "body=%s", rr.Body.String())

	var list []map[string]any
	list = testutil.DecodeJSON[[]map[string]any](t, rr)
	require.Len(t, list, 1)
	assert.Equal(t, w.Slug, list[0]["slug"])
}

func TestWorkspace_Create_Success(t *testing.T) {
	ts := testutil.NewTestServer(t)
	user := testutil.CreateUser(t, ts.DB)
	session := testutil.LoginAs(t, ts.DB, user)

	rr := ts.POST("/api/workspaces/", map[string]any{
		"name":              "Acme",
		"slug":              "acme-co",
		"organization_size": "10-50",
	}, session)
	require.Equal(t, http.StatusCreated, rr.Code, "body=%s", rr.Body.String())

	body := testutil.MustJSONMap(t, rr)
	assert.Equal(t, "Acme", body["name"])
	assert.Equal(t, "acme-co", body["slug"])
}

func TestWorkspace_Create_DuplicateSlug(t *testing.T) {
	ts := testutil.NewTestServer(t)
	user := testutil.CreateUser(t, ts.DB)
	_ = testutil.CreateWorkspace(t, ts.DB, user.ID) // factory uses ws-N slug
	session := testutil.LoginAs(t, ts.DB, user)

	// First create succeeds.
	rr := ts.POST("/api/workspaces/", map[string]any{
		"name": "First",
		"slug": "duplicate-slug",
	}, session)
	require.Equal(t, http.StatusCreated, rr.Code)

	// Second create with same slug → 409.
	rr2 := ts.POST("/api/workspaces/", map[string]any{
		"name": "Second",
		"slug": "duplicate-slug",
	}, session)
	require.Equal(t, http.StatusConflict, rr2.Code)
}

func TestWorkspace_Create_RequiresAuth(t *testing.T) {
	ts := testutil.NewTestServer(t)
	rr := ts.POST("/api/workspaces/", map[string]any{"name": "X"}, "")
	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestWorkspace_GetBySlug_Success(t *testing.T) {
	ts := testutil.NewTestServer(t)
	user := testutil.CreateUser(t, ts.DB)
	w := testutil.CreateWorkspace(t, ts.DB, user.ID)
	session := testutil.LoginAs(t, ts.DB, user)

	rr := ts.GET("/api/workspaces/"+w.Slug+"/", session)
	require.Equal(t, http.StatusOK, rr.Code, "body=%s", rr.Body.String())
	body := testutil.MustJSONMap(t, rr)
	assert.Equal(t, w.Slug, body["slug"])
}

func TestWorkspace_GetBySlug_NonMember404(t *testing.T) {
	ts := testutil.NewTestServer(t)
	owner := testutil.CreateUser(t, ts.DB)
	w := testutil.CreateWorkspace(t, ts.DB, owner.ID)

	stranger := testutil.CreateUser(t, ts.DB)
	session := testutil.LoginAs(t, ts.DB, stranger)

	rr := ts.GET("/api/workspaces/"+w.Slug+"/", session)
	require.Equal(t, http.StatusNotFound, rr.Code)
}

func TestWorkspace_Update_Name(t *testing.T) {
	ts := testutil.NewTestServer(t)
	user := testutil.CreateUser(t, ts.DB)
	w := testutil.CreateWorkspace(t, ts.DB, user.ID)
	session := testutil.LoginAs(t, ts.DB, user)

	rr := ts.PATCH("/api/workspaces/"+w.Slug+"/", map[string]any{
		"name": "New Name",
	}, session)
	require.Equal(t, http.StatusOK, rr.Code, "body=%s", rr.Body.String())
	assert.Equal(t, "New Name", testutil.MustJSONMap(t, rr)["name"])
}

func TestWorkspace_Delete_OwnerOnly(t *testing.T) {
	ts := testutil.NewTestServer(t)
	owner := testutil.CreateUser(t, ts.DB)
	w := testutil.CreateWorkspace(t, ts.DB, owner.ID)

	other := testutil.CreateUser(t, ts.DB)
	testutil.AddWorkspaceMember(t, ts.DB, w.ID, other.ID, testutil.RoleMember)
	otherSession := testutil.LoginAs(t, ts.DB, other)

	// Non-owner cannot delete.
	rr := ts.DELETE("/api/workspaces/"+w.Slug+"/", otherSession)
	assert.NotEqual(t, http.StatusNoContent, rr.Code, "non-owner should not be able to delete")

	// Owner can.
	ownerSession := testutil.LoginAs(t, ts.DB, owner)
	rr2 := ts.DELETE("/api/workspaces/"+w.Slug+"/", ownerSession)
	require.Equal(t, http.StatusNoContent, rr2.Code, "body=%s", rr2.Body.String())
}

func TestWorkspace_SlugCheck(t *testing.T) {
	ts := testutil.NewTestServer(t)
	user := testutil.CreateUser(t, ts.DB)
	w := testutil.CreateWorkspace(t, ts.DB, user.ID)
	session := testutil.LoginAs(t, ts.DB, user)

	// Available
	rr := ts.GET("/api/workspace-slug-check/?slug=brand-new-slug", session)
	require.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, true, testutil.MustJSONMap(t, rr)["status"])

	// Taken
	rr2 := ts.GET("/api/workspace-slug-check/?slug="+w.Slug, session)
	require.Equal(t, http.StatusOK, rr2.Code)
	assert.Equal(t, false, testutil.MustJSONMap(t, rr2)["status"])
}

func TestWorkspace_ListMembers(t *testing.T) {
	ts := testutil.NewTestServer(t)
	user := testutil.CreateUser(t, ts.DB)
	w := testutil.CreateWorkspace(t, ts.DB, user.ID)
	other := testutil.CreateUser(t, ts.DB)
	testutil.AddWorkspaceMember(t, ts.DB, w.ID, other.ID, testutil.RoleMember)
	session := testutil.LoginAs(t, ts.DB, user)

	rr := ts.GET("/api/workspaces/"+w.Slug+"/members/", session)
	require.Equal(t, http.StatusOK, rr.Code, "body=%s", rr.Body.String())
	list := testutil.DecodeJSON[[]map[string]any](t, rr)
	assert.Len(t, list, 2)
}

func TestWorkspace_UpdateMember_Role(t *testing.T) {
	ts := testutil.NewTestServer(t)
	owner := testutil.CreateUser(t, ts.DB)
	w := testutil.CreateWorkspace(t, ts.DB, owner.ID)
	other := testutil.CreateUser(t, ts.DB)
	m := testutil.AddWorkspaceMember(t, ts.DB, w.ID, other.ID, testutil.RoleMember)
	session := testutil.LoginAs(t, ts.DB, owner)

	rr := ts.PATCH("/api/workspaces/"+w.Slug+"/members/"+m.ID.String()+"/", map[string]any{
		"role": testutil.RoleAdmin,
	}, session)
	require.Equal(t, http.StatusOK, rr.Code, "body=%s", rr.Body.String())
	body := testutil.MustJSONMap(t, rr)
	assert.EqualValues(t, testutil.RoleAdmin, body["role"])
}

func TestWorkspace_DeleteMember(t *testing.T) {
	ts := testutil.NewTestServer(t)
	owner := testutil.CreateUser(t, ts.DB)
	w := testutil.CreateWorkspace(t, ts.DB, owner.ID)
	other := testutil.CreateUser(t, ts.DB)
	m := testutil.AddWorkspaceMember(t, ts.DB, w.ID, other.ID, testutil.RoleMember)
	session := testutil.LoginAs(t, ts.DB, owner)

	rr := ts.DELETE("/api/workspaces/"+w.Slug+"/members/"+m.ID.String()+"/", session)
	require.Equal(t, http.StatusNoContent, rr.Code, "body=%s", rr.Body.String())
}

func TestWorkspace_Leave_OwnerCannot(t *testing.T) {
	ts := testutil.NewTestServer(t)
	owner := testutil.CreateUser(t, ts.DB)
	w := testutil.CreateWorkspace(t, ts.DB, owner.ID)
	session := testutil.LoginAs(t, ts.DB, owner)

	rr := ts.POST("/api/workspaces/"+w.Slug+"/members/leave/", nil, session)
	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestWorkspace_Leave_MemberCan(t *testing.T) {
	ts := testutil.NewTestServer(t)
	owner := testutil.CreateUser(t, ts.DB)
	w := testutil.CreateWorkspace(t, ts.DB, owner.ID)
	other := testutil.CreateUser(t, ts.DB)
	testutil.AddWorkspaceMember(t, ts.DB, w.ID, other.ID, testutil.RoleMember)
	session := testutil.LoginAs(t, ts.DB, other)

	rr := ts.POST("/api/workspaces/"+w.Slug+"/members/leave/", nil, session)
	require.Equal(t, http.StatusNoContent, rr.Code, "body=%s", rr.Body.String())
}

func TestWorkspace_Invitations_CreateListDelete(t *testing.T) {
	ts := testutil.NewTestServer(t)
	owner := testutil.CreateUser(t, ts.DB)
	w := testutil.CreateWorkspace(t, ts.DB, owner.ID)
	session := testutil.LoginAs(t, ts.DB, owner)

	// Create
	rr := ts.POST("/api/workspaces/"+w.Slug+"/invitations/", map[string]any{
		"email": "newhire@test.local",
		"role":  testutil.RoleMember,
	}, session)
	require.Equal(t, http.StatusCreated, rr.Code, "body=%s", rr.Body.String())
	created := testutil.MustJSONMap(t, rr)
	inviteID, _ := created["id"].(string)
	require.NotEmpty(t, inviteID)

	// List
	rr2 := ts.GET("/api/workspaces/"+w.Slug+"/invitations/", session)
	require.Equal(t, http.StatusOK, rr2.Code)
	list := testutil.DecodeJSON[[]map[string]any](t, rr2)
	assert.Len(t, list, 1)

	// Get one
	rr3 := ts.GET("/api/workspaces/"+w.Slug+"/invitations/"+inviteID+"/", session)
	require.Equal(t, http.StatusOK, rr3.Code)

	// Delete
	rr4 := ts.DELETE("/api/workspaces/"+w.Slug+"/invitations/"+inviteID+"/", session)
	require.Equal(t, http.StatusNoContent, rr4.Code)
}

func TestWorkspace_JoinByToken(t *testing.T) {
	ts := testutil.NewTestServer(t)
	owner := testutil.CreateUser(t, ts.DB)
	w := testutil.CreateWorkspace(t, ts.DB, owner.ID)
	invitee := testutil.CreateUser(t, ts.DB, testutil.WithUserEmail("joinme@test.local"))
	inv := testutil.CreateWorkspaceInvite(t, ts.DB, w.ID, "joinme@test.local", "join-tok-123")

	session := testutil.LoginAs(t, ts.DB, invitee)
	rr := ts.POST("/api/workspaces/join/", map[string]any{"token": inv.Token}, session)
	require.Equal(t, http.StatusOK, rr.Code, "body=%s", rr.Body.String())
	body := testutil.MustJSONMap(t, rr)
	assert.Equal(t, w.Slug, body["slug"])
}

func TestWorkspace_JoinByInvite_PathID(t *testing.T) {
	ts := testutil.NewTestServer(t)
	owner := testutil.CreateUser(t, ts.DB)
	w := testutil.CreateWorkspace(t, ts.DB, owner.ID)
	invitee := testutil.CreateUser(t, ts.DB, testutil.WithUserEmail("byid@test.local"))
	inv := testutil.CreateWorkspaceInvite(t, ts.DB, w.ID, "byid@test.local", "byid-tok-456")

	session := testutil.LoginAs(t, ts.DB, invitee)
	rr := ts.POST("/api/workspaces/"+w.Slug+"/invitations/"+inv.ID.String()+"/join/", nil, session)
	require.Equal(t, http.StatusOK, rr.Code, "body=%s", rr.Body.String())
}

func TestWorkspace_ListUserInvitations(t *testing.T) {
	ts := testutil.NewTestServer(t)
	owner := testutil.CreateUser(t, ts.DB)
	w := testutil.CreateWorkspace(t, ts.DB, owner.ID)
	user := testutil.CreateUser(t, ts.DB, testutil.WithUserEmail("inv@test.local"))
	_ = testutil.CreateWorkspaceInvite(t, ts.DB, w.ID, "inv@test.local", "inv-tok-789")

	session := testutil.LoginAs(t, ts.DB, user)
	rr := ts.GET("/api/users/me/workspaces/invitations/", session)
	require.Equal(t, http.StatusOK, rr.Code, "body=%s", rr.Body.String())
}
