package handler_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/Devlaner/devlane/api/internal/model"
	"github.com/Devlaner/devlane/api/internal/testutil"
	"github.com/stretchr/testify/require"
)

// TestWorkspace_JoinByToken_RejectsEmailMismatch proves a workspace invite
// token can't be redeemed by an account whose email doesn't match the
// invited email (a leaked/forwarded token must not grant membership).
func TestWorkspace_JoinByToken_RejectsEmailMismatch(t *testing.T) {
	ts := testutil.NewTestServer(t)
	owner := testutil.CreateUser(t, ts.DB)
	w := testutil.CreateWorkspace(t, ts.DB, owner.ID)
	inv := testutil.CreateWorkspaceInvite(t, ts.DB, w.ID, "invited@test.local", "tok-mismatch-111")

	stranger := testutil.CreateUser(t, ts.DB) // email is stranger-N@test.local, not invited@test.local
	session := testutil.LoginAs(t, ts.DB, stranger)

	rr := ts.POST("/api/workspaces/join/", map[string]any{"token": inv.Token}, session)
	require.Equal(t, http.StatusNotFound, rr.Code, "body=%s", rr.Body.String())
}

// TestWorkspace_JoinByToken_AcceptsEmailMatch proves the happy path still
// works when the joining account's email matches the invite.
func TestWorkspace_JoinByToken_AcceptsEmailMatch(t *testing.T) {
	ts := testutil.NewTestServer(t)
	owner := testutil.CreateUser(t, ts.DB)
	w := testutil.CreateWorkspace(t, ts.DB, owner.ID)

	invitee := testutil.CreateUser(t, ts.DB)
	email := invitee.Email
	require.NotNil(t, email)
	inv := testutil.CreateWorkspaceInvite(t, ts.DB, w.ID, *email, "tok-match-111")
	session := testutil.LoginAs(t, ts.DB, invitee)

	rr := ts.POST("/api/workspaces/join/", map[string]any{"token": inv.Token}, session)
	require.Equal(t, http.StatusOK, rr.Code, "body=%s", rr.Body.String())
}

// TestWorkspace_JoinByToken_AllowsRejoinAfterRemoval proves a user who was
// previously removed from a workspace (leaving a soft-deleted membership
// row) can rejoin via a fresh invite, instead of hitting the
// UNIQUE(workspace_id, member_id) constraint against the old row.
func TestWorkspace_JoinByToken_AllowsRejoinAfterRemoval(t *testing.T) {
	ts := testutil.NewTestServer(t)
	owner := testutil.CreateUser(t, ts.DB)
	w := testutil.CreateWorkspace(t, ts.DB, owner.ID)

	rejoiner := testutil.CreateUser(t, ts.DB)
	testutil.AddWorkspaceMember(t, ts.DB, w.ID, rejoiner.ID, testutil.RoleMember)
	require.NoError(t, ts.DB.WithContext(context.Background()).
		Where("workspace_id = ? AND member_id = ?", w.ID, rejoiner.ID).
		Delete(&model.WorkspaceMember{}).Error)

	email := rejoiner.Email
	require.NotNil(t, email)
	inv := testutil.CreateWorkspaceInvite(t, ts.DB, w.ID, *email, "tok-rejoin-111")
	session := testutil.LoginAs(t, ts.DB, rejoiner)

	rr := ts.POST("/api/workspaces/join/", map[string]any{"token": inv.Token}, session)
	require.Equal(t, http.StatusOK, rr.Code, "body=%s", rr.Body.String())
}

func TestProject_JoinByToken_RejectsEmailMismatch(t *testing.T) {
	ts := testutil.NewTestServer(t)
	owner := testutil.CreateUser(t, ts.DB)
	w := testutil.CreateWorkspace(t, ts.DB, owner.ID)
	p := testutil.CreateProject(t, ts.DB, w.ID, owner.ID)

	inv := &model.ProjectMemberInvite{
		ProjectID:   p.ID,
		WorkspaceID: w.ID,
		Email:       "invited@test.local",
		Token:       "proj-tok-mismatch-111",
		Role:        testutil.RoleMember,
	}
	require.NoError(t, ts.DB.WithContext(context.Background()).Create(inv).Error)

	stranger := testutil.CreateUser(t, ts.DB)
	session := testutil.LoginAs(t, ts.DB, stranger)

	rr := ts.POST("/api/workspaces/"+w.Slug+"/projects/join/", map[string]any{"token": inv.Token}, session)
	require.Equal(t, http.StatusNotFound, rr.Code, "body=%s", rr.Body.String())
}

func TestProject_JoinByToken_AcceptsEmailMatch(t *testing.T) {
	ts := testutil.NewTestServer(t)
	owner := testutil.CreateUser(t, ts.DB)
	w := testutil.CreateWorkspace(t, ts.DB, owner.ID)
	p := testutil.CreateProject(t, ts.DB, w.ID, owner.ID)

	invitee := testutil.CreateUser(t, ts.DB)
	testutil.AddWorkspaceMember(t, ts.DB, w.ID, invitee.ID, testutil.RoleMember)
	email := invitee.Email
	require.NotNil(t, email)

	inv := &model.ProjectMemberInvite{
		ProjectID:   p.ID,
		WorkspaceID: w.ID,
		Email:       *email,
		Token:       "proj-tok-match-111",
		Role:        testutil.RoleMember,
	}
	require.NoError(t, ts.DB.WithContext(context.Background()).Create(inv).Error)

	session := testutil.LoginAs(t, ts.DB, invitee)
	rr := ts.POST("/api/workspaces/"+w.Slug+"/projects/join/", map[string]any{"token": inv.Token}, session)
	require.Equal(t, http.StatusOK, rr.Code, "body=%s", rr.Body.String())
}

// TestProject_JoinByToken_AllowsRejoinAfterRemoval mirrors the workspace
// version: a soft-deleted project_members row for the same (project, member)
// must not block rejoining via a fresh invite.
func TestProject_JoinByToken_AllowsRejoinAfterRemoval(t *testing.T) {
	ts := testutil.NewTestServer(t)
	owner := testutil.CreateUser(t, ts.DB)
	w := testutil.CreateWorkspace(t, ts.DB, owner.ID)
	p := testutil.CreateProject(t, ts.DB, w.ID, owner.ID)

	rejoiner := testutil.CreateUser(t, ts.DB)
	testutil.AddWorkspaceMember(t, ts.DB, w.ID, rejoiner.ID, testutil.RoleMember)
	testutil.AddProjectMember(t, ts.DB, p.ID, w.ID, rejoiner.ID, testutil.RoleMember)
	require.NoError(t, ts.DB.WithContext(context.Background()).
		Where("project_id = ? AND member_id = ?", p.ID, rejoiner.ID).
		Delete(&model.ProjectMember{}).Error)

	email := rejoiner.Email
	require.NotNil(t, email)
	inv := &model.ProjectMemberInvite{
		ProjectID:   p.ID,
		WorkspaceID: w.ID,
		Email:       *email,
		Token:       "proj-tok-rejoin-111",
		Role:        testutil.RoleMember,
	}
	require.NoError(t, ts.DB.WithContext(context.Background()).Create(inv).Error)

	session := testutil.LoginAs(t, ts.DB, rejoiner)
	rr := ts.POST("/api/workspaces/"+w.Slug+"/projects/join/", map[string]any{"token": inv.Token}, session)
	require.Equal(t, http.StatusOK, rr.Code, "body=%s", rr.Body.String())
}
