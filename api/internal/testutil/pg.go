// Package testutil provides shared infrastructure for HTTP-level API tests.
//
// Usage in a test file:
//
//	func TestSomething(t *testing.T) {
//	    ts := testutil.NewTestServer(t)
//	    user := testutil.CreateUser(t, ts.DB)
//	    sessionKey := testutil.LoginAs(t, ts.DB, user)
//	    rr := ts.GET("/api/users/me/", sessionKey)
//	    require.Equal(t, http.StatusOK, rr.Code)
//	}
//
// One Postgres testcontainer is started per `go test` package invocation
// (lazy, behind sync.Once). Tests share the same database; NewTestServer
// truncates every table at the start of each test for isolation.
package testutil

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/Devlaner/devlane/api/internal/config"
	"github.com/Devlaner/devlane/api/internal/database"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	pgOnce sync.Once
	pgDB   *gorm.DB
	pgErr  error
)

// PG returns the shared test Postgres database, starting the container on
// first call. Migrations run once. Subsequent calls reuse the connection.
func PG(t testing.TB) *gorm.DB {
	t.Helper()
	pgOnce.Do(initPG)
	if pgErr != nil {
		t.Fatalf("postgres testcontainer: %v", pgErr)
	}
	return pgDB
}

func initPG() {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("devlane_test"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(90*time.Second),
		),
	)
	if err != nil {
		pgErr = fmt.Errorf("start container: %w", err)
		return
	}

	host, err := container.Host(ctx)
	if err != nil {
		pgErr = fmt.Errorf("container host: %w", err)
		return
	}
	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		pgErr = fmt.Errorf("container port: %w", err)
		return
	}

	cfg := &config.Config{
		DBHost:         host,
		DBPort:         port.Port(),
		DBUser:         "postgres",
		DBPassword:     "postgres",
		DBName:         "devlane_test",
		DBSSLMode:      "disable",
		MigrationsPath: migrationsPath(),
	}

	silentLog := slog.New(slog.NewTextHandler(discardWriter{}, &slog.HandlerOptions{Level: slog.LevelError}))
	if err := database.RunMigrations(cfg, silentLog); err != nil {
		pgErr = fmt.Errorf("run migrations: %w", err)
		return
	}

	dsn := cfg.DSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		pgErr = fmt.Errorf("open gorm: %w", err)
		return
	}

	pgDB = db
}

// migrationsPath returns the absolute path to api/migrations/ derived from
// this file's source location, so tests work regardless of CWD. Returned
// with forward slashes — golang-migrate's file source prefixes "file://"
// and then strips it, so "C:/path" (Windows) and "/path" (Unix) both work.
func migrationsPath() string {
	_, file, _, _ := runtime.Caller(0)
	// file = .../api/internal/testutil/pg.go
	apiDir := filepath.Join(filepath.Dir(file), "..", "..")
	abs, err := filepath.Abs(filepath.Join(apiDir, "migrations"))
	if err != nil {
		abs = filepath.Join(apiDir, "migrations")
	}
	return filepath.ToSlash(abs)
}

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }
