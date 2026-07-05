package handler_test

import (
	"context"
	"testing"

	"github.com/Devlaner/devlane/api/internal/model"
	"github.com/Devlaner/devlane/api/internal/service"
	"github.com/Devlaner/devlane/api/internal/store"
	"github.com/Devlaner/devlane/api/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// A leaked attachment object URL must not be fetchable by a non-member.
func TestAttachment_AuthorizeDownload(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)
	issue := testutil.CreateIssue(t, ts.DB, w.Project.ID, w.Workspace.ID, w.User.ID)

	assetID := uuid.New()
	require.NoError(t, ts.DB.Create(&model.FileAsset{
		ID:      assetID,
		Asset:   "attachments/" + issue.ID.String() + "/" + assetID.String(),
		IssueID: &issue.ID,
	}).Error)
	require.NoError(t, ts.DB.Create(&model.IssueAttachment{
		ID:          uuid.New(),
		IssueID:     issue.ID,
		AssetID:     assetID,
		ProjectID:   w.Project.ID,
		WorkspaceID: w.Workspace.ID,
	}).Error)

	svc := service.NewAttachmentService(
		store.NewIssueStore(ts.DB),
		store.NewProjectStore(ts.DB),
		store.NewWorkspaceStore(ts.DB),
		nil,
	)
	ctx := context.Background()

	// A workspace member (the owner) may download.
	require.NoError(t, svc.AuthorizeDownload(ctx, issue.ID, assetID, w.User.ID))

	// A user who isn't in the workspace is denied.
	outsider := testutil.CreateUser(t, ts.DB)
	require.Error(t, svc.AuthorizeDownload(ctx, issue.ID, assetID, outsider.ID))

	// A non-existent attachment (wrong asset id) is not found.
	require.Error(t, svc.AuthorizeDownload(ctx, issue.ID, uuid.New(), w.User.ID))
}
