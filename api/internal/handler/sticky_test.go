package handler_test

import (
	"net/http"
	"testing"

	"github.com/Devlaner/devlane/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSticky_RequiresAuth(t *testing.T) {
	ts := testutil.NewTestServer(t)
	rr := ts.GET("/api/workspaces/x/stickies/", "")
	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestSticky_CRUD(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)
	base := "/api/workspaces/" + w.Workspace.Slug + "/stickies/"

	// Create
	rr := ts.POST(base, map[string]any{
		"name":        "Reminder",
		"description": "Don't forget the deploy",
		"color":       "#ffeb3b",
	}, w.Session)
	require.Equal(t, http.StatusCreated, rr.Code, "body=%s", rr.Body.String())
	id, _ := testutil.MustJSONMap(t, rr)["id"].(string)
	require.NotEmpty(t, id)

	// List
	rr2 := ts.GET(base, w.Session)
	require.Equal(t, http.StatusOK, rr2.Code)
	list := testutil.DecodeJSON[[]map[string]any](t, rr2)
	assert.Len(t, list, 1)

	// Update
	rr3 := ts.PATCH(base+id+"/", map[string]any{"name": "Updated reminder"}, w.Session)
	require.Equal(t, http.StatusOK, rr3.Code, "body=%s", rr3.Body.String())

	// Delete
	rr4 := ts.DELETE(base+id+"/", w.Session)
	require.Equal(t, http.StatusNoContent, rr4.Code)
}
