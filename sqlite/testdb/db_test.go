package testdb

import "testing"

func TestMigrationsCreateSchema(t *testing.T) {
	db := Open(t)

	tables := []string{"user", "location", "event", "event_result", "friend"}
	for _, table := range tables {
		var name string
		err := db.QueryRow(
			`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`,
			table,
		).Scan(&name)
		if err != nil {
			t.Fatalf("table %s: %v", table, err)
		}
	}
}

func TestDatabasesAreIndependent(t *testing.T) {
	first := Open(t)
	second := Open(t)

	if _, err := first.Exec(`INSERT INTO location (name) VALUES ('Alpha')`); err != nil {
		t.Fatalf("insert location: %v", err)
	}

	var count int
	if err := second.QueryRow(`SELECT COUNT(*) FROM location`).Scan(&count); err != nil {
		t.Fatalf("count locations: %v", err)
	}
	if count != 0 {
		t.Fatalf("second database has %d locations, want 0", count)
	}
}
