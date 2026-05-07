package testutil

import (
	"strings"
	"sync"
	"testing"

	"gorm.io/gorm"
)

var (
	tableListMu  sync.Mutex
	cachedTables []string
)

// TruncateAll empties every public-schema table except schema_migrations.
// Uses a single TRUNCATE … RESTART IDENTITY CASCADE statement so foreign
// key constraints don't get in the way of arbitrary truncation order.
func TruncateAll(t testing.TB, db *gorm.DB) {
	t.Helper()

	tables, err := publicTables(db)
	if err != nil {
		t.Fatalf("list tables: %v", err)
	}
	if len(tables) == 0 {
		return
	}

	quoted := make([]string, 0, len(tables))
	for _, name := range tables {
		quoted = append(quoted, `"`+name+`"`)
	}
	stmt := "TRUNCATE TABLE " + strings.Join(quoted, ", ") + " RESTART IDENTITY CASCADE"
	if err := db.Exec(stmt).Error; err != nil {
		t.Fatalf("truncate: %v", err)
	}
}

func publicTables(db *gorm.DB) ([]string, error) {
	tableListMu.Lock()
	defer tableListMu.Unlock()

	if cachedTables != nil {
		return cachedTables, nil
	}

	var tables []string
	err := db.Raw(`
		SELECT tablename
		FROM pg_tables
		WHERE schemaname = 'public'
		  AND tablename != 'schema_migrations'
		ORDER BY tablename
	`).Scan(&tables).Error
	if err != nil {
		return nil, err
	}
	cachedTables = tables
	return tables, nil
}
