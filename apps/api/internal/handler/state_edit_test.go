package handler_test

import (
	"net/http"
	"testing"

	"github.com/Devlaner/devlane/api/internal/model"
	"github.com/Devlaner/devlane/api/internal/testutil"
	"github.com/stretchr/testify/require"
)

// A custom state's group, sequence, and default flag are all editable, and
// making one state the default clears the flag on the previous default.
func TestState_UpdateGroupSequenceAndDefault(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)
	s1 := testutil.CreateState(t, ts.DB, w.Project.ID, w.Workspace.ID)
	s2 := testutil.CreateState(t, ts.DB, w.Project.ID, w.Workspace.ID)
	base := "/api/workspaces/" + w.Workspace.Slug + "/projects/" + w.Project.ID.String() + "/states/"

	// Make s1 the default.
	require.Equal(t, http.StatusOK,
		ts.PATCH(base+s1.ID.String()+"/", map[string]any{"default": true}, w.Session).Code)

	// Change s2's group and sequence.
	require.Equal(t, http.StatusOK,
		ts.PATCH(base+s2.ID.String()+"/", map[string]any{"group": "started", "sequence": 5}, w.Session).Code)

	// A group outside the accepted set is rejected.
	require.Equal(t, http.StatusBadRequest,
		ts.PATCH(base+s1.ID.String()+"/", map[string]any{"group": "bogus"}, w.Session).Code)

	// Making s2 the default must clear s1's default.
	require.Equal(t, http.StatusOK,
		ts.PATCH(base+s2.ID.String()+"/", map[string]any{"default": true}, w.Session).Code)

	var got1, got2 model.State
	require.NoError(t, ts.DB.First(&got1, "id = ?", s1.ID).Error)
	require.NoError(t, ts.DB.First(&got2, "id = ?", s2.ID).Error)
	require.False(t, got1.Default, "previous default should be cleared")
	require.True(t, got2.Default, "new default should be set")
	require.Equal(t, "started", got2.Group)
	require.Equal(t, float64(5), got2.Sequence)
}

// Reorder assigns new sequences to several states atomically in one request.
func TestState_ReorderSequences(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)
	s1 := testutil.CreateState(t, ts.DB, w.Project.ID, w.Workspace.ID)
	s2 := testutil.CreateState(t, ts.DB, w.Project.ID, w.Workspace.ID)
	s3 := testutil.CreateState(t, ts.DB, w.Project.ID, w.Workspace.ID)
	base := "/api/workspaces/" + w.Workspace.Slug + "/projects/" + w.Project.ID.String() + "/states/"

	body := map[string]any{"states": []map[string]any{
		{"id": s3.ID.String(), "sequence": 0},
		{"id": s1.ID.String(), "sequence": 1},
		{"id": s2.ID.String(), "sequence": 2},
	}}
	require.Equal(t, http.StatusNoContent, ts.POST(base+"reorder/", body, w.Session).Code)

	var g1, g2, g3 model.State
	require.NoError(t, ts.DB.First(&g1, "id = ?", s1.ID).Error)
	require.NoError(t, ts.DB.First(&g2, "id = ?", s2.ID).Error)
	require.NoError(t, ts.DB.First(&g3, "id = ?", s3.ID).Error)
	require.Equal(t, float64(0), g3.Sequence)
	require.Equal(t, float64(1), g1.Sequence)
	require.Equal(t, float64(2), g2.Sequence)
}
