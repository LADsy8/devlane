package handler_test

import (
	"net/http"
	"testing"

	"github.com/Devlaner/devlane/api/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestUpload_RequiresAuth_Upload(t *testing.T) {
	ts := testutil.NewTestServer(t)
	rr := ts.POST("/api/upload", nil, "")
	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestUpload_NoMinioReturns503(t *testing.T) {
	// The TestServer is constructed with Minio: nil — once authenticated, the
	// handler short-circuits with 503 ("feature unconfigured") instead of
	// attempting to use the nil client.
	ts := testutil.NewTestServer(t)
	user := testutil.CreateUser(t, ts.DB)
	session := testutil.LoginAs(t, ts.DB, user)

	rr := ts.POST("/api/upload", nil, session)
	require.Equal(t, http.StatusServiceUnavailable, rr.Code)
}

func TestUpload_ServeFile_NoMinioReturns503(t *testing.T) {
	ts := testutil.NewTestServer(t)
	user := testutil.CreateUser(t, ts.DB)
	session := testutil.LoginAs(t, ts.DB, user)

	rr := ts.GET("/api/files/uploads/2026/05/abc.png", session)
	require.Equal(t, http.StatusServiceUnavailable, rr.Code)
}

func TestUpload_RequiresAuth(t *testing.T) {
	ts := testutil.NewTestServer(t)
	rr := ts.GET("/api/files/uploads/anything", "")
	require.Equal(t, http.StatusUnauthorized, rr.Code)
}
