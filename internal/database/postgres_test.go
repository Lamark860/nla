package database

import (
	"strings"
	"testing"
)

// Guards the embed contract: if someone moves migrations out of
// internal/database/migrations/, RunMigrations would silently start applying nothing.
func TestMigrationsEmbedded(t *testing.T) {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		t.Fatalf("read migrations dir: %v", err)
	}

	var sqlFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			sqlFiles = append(sqlFiles, e.Name())
		}
	}

	if len(sqlFiles) == 0 {
		t.Fatal("no .sql migration files embedded")
	}

	for _, name := range sqlFiles {
		body, err := migrationsFS.ReadFile("migrations/" + name)
		if err != nil {
			t.Errorf("read %s: %v", name, err)
			continue
		}
		if len(strings.TrimSpace(string(body))) == 0 {
			t.Errorf("%s is empty", name)
		}
	}
}
