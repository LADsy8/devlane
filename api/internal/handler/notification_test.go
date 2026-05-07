package handler_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/Devlaner/devlane/api/internal/model"
	"github.com/Devlaner/devlane/api/internal/store"
	"github.com/Devlaner/devlane/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotification_RequiresAuth(t *testing.T) {
	ts := testutil.NewTestServer(t)
	rr := ts.GET("/api/workspaces/x/notifications/", "")
	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestNotification_List_Empty(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)

	rr := ts.GET("/api/workspaces/"+w.Workspace.Slug+"/notifications/", w.Session)
	require.Equal(t, http.StatusOK, rr.Code, "body=%s", rr.Body.String())
}

func TestNotification_UnreadCount(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)

	rr := ts.GET("/api/workspaces/"+w.Workspace.Slug+"/notifications/unread-count/", w.Session)
	require.Equal(t, http.StatusOK, rr.Code, "body=%s", rr.Body.String())
	body := testutil.MustJSONMap(t, rr)
	assert.EqualValues(t, 0, body["total"])
}

func TestNotification_MarkAllRead_NoOp(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)

	rr := ts.POST("/api/workspaces/"+w.Workspace.Slug+"/notifications/mark-all-read/", nil, w.Session)
	require.Truef(t, rr.Code < 400, "body=%s", rr.Body.String())
}

func TestNotification_LifecycleWithSeed(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)

	triggeredBy := w.User.ID
	n := &model.Notification{
		WorkspaceID:   w.Workspace.ID,
		ProjectID:     &w.Project.ID,
		ReceiverID:    w.User.ID,
		TriggeredByID: &triggeredBy,
		Title:         "Test notification",
		Message:       model.JSONMap{"text": "Hello"},
		EntityName:    model.NotificationEntityIssue,
		Sender:        model.NotificationSenderSubscribed,
	}
	require.NoError(t, store.NewNotificationStore(ts.DB).Create(context.Background(), n))

	// List (now has 1)
	rr := ts.GET("/api/workspaces/"+w.Workspace.Slug+"/notifications/", w.Session)
	require.Equal(t, http.StatusOK, rr.Code, "body=%s", rr.Body.String())
	list := testutil.DecodeJSON[[]map[string]any](t, rr)
	require.GreaterOrEqual(t, len(list), 1)

	// MarkRead
	rr2 := ts.POST("/api/workspaces/"+w.Workspace.Slug+"/notifications/"+n.ID.String()+"/read/", nil, w.Session)
	require.Truef(t, rr2.Code < 400, "body=%s", rr2.Body.String())

	// MarkUnread
	rr3 := ts.DELETE("/api/workspaces/"+w.Workspace.Slug+"/notifications/"+n.ID.String()+"/read/", w.Session)
	require.Truef(t, rr3.Code < 400, "body=%s", rr3.Body.String())

	// Archive
	rr4 := ts.POST("/api/workspaces/"+w.Workspace.Slug+"/notifications/"+n.ID.String()+"/archive/", nil, w.Session)
	require.Truef(t, rr4.Code < 400, "body=%s", rr4.Body.String())

	// Unarchive
	rr5 := ts.DELETE("/api/workspaces/"+w.Workspace.Slug+"/notifications/"+n.ID.String()+"/archive/", w.Session)
	require.Truef(t, rr5.Code < 400, "body=%s", rr5.Body.String())

	// Snooze (until tomorrow)
	rr6 := ts.POST("/api/workspaces/"+w.Workspace.Slug+"/notifications/"+n.ID.String()+"/snooze/",
		map[string]any{"until": time.Now().Add(24 * time.Hour).Format(time.RFC3339)}, w.Session)
	require.Truef(t, rr6.Code < 400, "body=%s", rr6.Body.String())

	// Unsnooze
	rr7 := ts.DELETE("/api/workspaces/"+w.Workspace.Slug+"/notifications/"+n.ID.String()+"/snooze/", w.Session)
	require.Truef(t, rr7.Code < 400, "body=%s", rr7.Body.String())
}
