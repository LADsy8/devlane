package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/Devlaner/devlane/api/internal/model"
	"github.com/Devlaner/devlane/api/internal/store"
	"github.com/Devlaner/devlane/api/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// Two writers hold the same snapshot and each change a different field. With a
// column-scoped update, both changes survive; a full-row Save from either stale
// snapshot would revert the other's field (the lost-update bug in #144).
func TestIssueStore_UpdateFields_DoesNotClobberConcurrentColumn(t *testing.T) {
	db := testutil.NewTestServer(t).DB
	world := testutil.SeedWorld(t, db)
	is := store.NewIssueStore(db)
	ctx := context.Background()

	issue := testutil.CreateIssue(t, db, world.Project.ID, world.Workspace.ID, world.User.ID)
	originalName := issue.Name

	// Writer B changes the name.
	require.NoError(t, is.UpdateFields(ctx, issue.ID, map[string]any{"name": "changed-by-B"}))

	// Writer A, still holding the pre-B snapshot, changes only the priority.
	require.NoError(t, is.UpdateFields(ctx, issue.ID, map[string]any{"priority": "urgent"}))

	var got model.Issue
	require.NoError(t, db.First(&got, "id = ?", issue.ID).Error)
	require.Equal(t, "urgent", got.Priority, "writer A's change should apply")
	require.Equal(t, "changed-by-B", got.Name, "writer B's change must not be reverted")
	require.NotEqual(t, originalName, got.Name)
}

// A nil pointer in the update map clears the column (the service relies on this
// to clear start/target dates and the estimate).
func TestIssueStore_UpdateFields_NilClearsColumn(t *testing.T) {
	db := testutil.NewTestServer(t).DB
	world := testutil.SeedWorld(t, db)
	is := store.NewIssueStore(db)
	ctx := context.Background()

	issue := testutil.CreateIssue(t, db, world.Project.ID, world.Workspace.ID, world.User.ID)

	due := time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC)
	require.NoError(t, is.UpdateFields(ctx, issue.ID, map[string]any{"target_date": &due}))
	var withDate model.Issue
	require.NoError(t, db.First(&withDate, "id = ?", issue.ID).Error)
	require.NotNil(t, withDate.TargetDate)

	var cleared *time.Time
	require.NoError(t, is.UpdateFields(ctx, issue.ID, map[string]any{"target_date": cleared}))
	var afterClear model.Issue
	require.NoError(t, db.First(&afterClear, "id = ?", issue.ID).Error)
	require.Nil(t, afterClear.TargetDate)
}

// Updating a non-existent (or already-deleted) issue must not silently succeed.
func TestIssueStore_UpdateFields_ReturnsNotFoundWhenMissing(t *testing.T) {
	db := testutil.NewTestServer(t).DB
	is := store.NewIssueStore(db)

	err := is.UpdateFields(context.Background(), uuid.New(), map[string]any{"priority": "urgent"})
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
