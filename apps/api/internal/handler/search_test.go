package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/Devlaner/devlane/api/internal/model"
	"github.com/Devlaner/devlane/api/internal/testutil"
	"github.com/stretchr/testify/require"
)

type searchResponse struct {
	Results struct {
		Issue   []map[string]any `json:"issue"`
		Epic    []map[string]any `json:"epic"`
		Cycle   []map[string]any `json:"cycle"`
		Module  []map[string]any `json:"module"`
		View    []map[string]any `json:"view"`
		Page    []map[string]any `json:"page"`
		Project []map[string]any `json:"project"`
	} `json:"results"`
}

func parseSearch(t *testing.T, rr *httptest.ResponseRecorder) searchResponse {
	t.Helper()
	var resp searchResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp), "body=%s", rr.Body.String())
	return resp
}

func names(rows []map[string]any) []string {
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		if n, ok := r["name"].(string); ok {
			out = append(out, n)
		}
	}
	return out
}

func TestSearch_AcrossEntities(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)

	// Seed one of each entity with a distinctive name.
	issue := testutil.CreateIssue(t, ts.DB, w.Project.ID, w.Workspace.ID, w.User.ID)
	require.NoError(t, ts.DB.Model(issue).Update("name", "Login flow bug").Error)

	require.NoError(t, ts.DB.Create(&model.Cycle{
		Name: "Sprint Alpha", ProjectID: w.Project.ID, WorkspaceID: w.Workspace.ID,
		OwnedByID: w.User.ID, CreatedByID: &w.User.ID,
	}).Error)
	require.NoError(t, ts.DB.Create(&model.Module{
		Name: "Payments service", ProjectID: w.Project.ID, WorkspaceID: w.Workspace.ID,
		CreatedByID: &w.User.ID,
	}).Error)
	require.NoError(t, ts.DB.Create(&model.View{
		Name: "My open items", ProjectID: w.Project.ID, WorkspaceID: w.Workspace.ID,
		Query: model.JSONMap{}, CreatedByID: &w.User.ID,
	}).Error)

	base := "/api/workspaces/" + w.Workspace.Slug + "/search/"

	// Match the issue by name; nothing else should match "login".
	r1 := ts.GET(base+"?query=login", w.Session)
	require.Equal(t, http.StatusOK, r1.Code, "body=%s", r1.Body.String())
	resp1 := parseSearch(t, r1)
	require.Contains(t, names(resp1.Results.Issue), "Login flow bug")
	require.Empty(t, resp1.Results.Cycle)
	require.Empty(t, resp1.Results.Module)

	// Match the cycle.
	resp2 := parseSearch(t, ts.GET(base+"?query=sprint", w.Session))
	require.Contains(t, names(resp2.Results.Cycle), "Sprint Alpha")

	// Match the module.
	resp3 := parseSearch(t, ts.GET(base+"?query=payments", w.Session))
	require.Contains(t, names(resp3.Results.Module), "Payments service")

	// Match the view.
	resp4 := parseSearch(t, ts.GET(base+"?query=open+items", w.Session))
	require.Contains(t, names(resp4.Results.View), "My open items")

	// Match the seeded project by its name.
	resp5 := parseSearch(t, ts.GET(base+"?query="+url.QueryEscape(w.Project.Name), w.Session))
	require.Contains(t, names(resp5.Results.Project), w.Project.Name)
}

func TestSearch_BlankQueryReturnsEmpty(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)
	issue := testutil.CreateIssue(t, ts.DB, w.Project.ID, w.Workspace.ID, w.User.ID)
	require.NoError(t, ts.DB.Model(issue).Update("name", "Findable").Error)

	base := "/api/workspaces/" + w.Workspace.Slug + "/search/"
	rr := ts.GET(base+"?query=%20%20", w.Session)
	require.Equal(t, http.StatusOK, rr.Code, "body=%s", rr.Body.String())
	resp := parseSearch(t, rr)
	require.Empty(t, resp.Results.Issue)
	require.Empty(t, resp.Results.Project)
}

func TestSearch_NonMemberForbidden(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)

	// A second user who is not a member of w.Workspace.
	outsider := testutil.CreateUser(t, ts.DB)
	session := testutil.LoginAs(t, ts.DB, outsider)

	base := "/api/workspaces/" + w.Workspace.Slug + "/search/?query=anything"
	rr := ts.GET(base, session)
	require.Equal(t, http.StatusForbidden, rr.Code, "body=%s", rr.Body.String())
}
