package handler_test

import (
	"net/http"
	"sync"
	"testing"

	"github.com/Devlaner/devlane/api/internal/model"
	"github.com/Devlaner/devlane/api/internal/testutil"
	"github.com/stretchr/testify/require"
)

// Two PATCHes to different fields of the same issue, fired together, must both
// persist. Before the fix the update did a full-row Save from an in-memory
// snapshot, so whichever request committed last reverted the other's field.
func TestIssue_ConcurrentUpdatesToDifferentFields_BothPersist(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)

	for i := 0; i < 5; i++ {
		issue := testutil.CreateIssue(t, ts.DB, w.Project.ID, w.Workspace.ID, w.User.ID)
		base := "/api/workspaces/" + w.Workspace.Slug + "/projects/" + w.Project.ID.String() +
			"/issues/" + issue.ID.String() + "/"

		start := make(chan struct{})
		var wg sync.WaitGroup
		var priCode, nameCode int
		wg.Add(2)
		go func() {
			defer wg.Done()
			<-start
			priCode = ts.PATCH(base, map[string]any{"priority": "urgent"}, w.Session).Code
		}()
		go func() {
			defer wg.Done()
			<-start
			nameCode = ts.PATCH(base, map[string]any{"name": "renamed"}, w.Session).Code
		}()
		close(start) // release both at once to maximise interleaving
		wg.Wait()

		require.Equal(t, http.StatusOK, priCode)
		require.Equal(t, http.StatusOK, nameCode)

		var got model.Issue
		require.NoError(t, ts.DB.First(&got, "id = ?", issue.ID).Error)
		require.Equalf(t, "urgent", got.Priority, "priority change lost on iteration %d", i)
		require.Equalf(t, "renamed", got.Name, "name change lost on iteration %d", i)
	}
}
