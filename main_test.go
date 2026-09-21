package main

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/duckdb/duckdb-go/v2"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("duckdb", "")
	if err != nil {
		t.Fatalf("open DuckDB: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestDuckDBVersionIsPinnedToLTS(t *testing.T) {
	var version string
	if err := openTestDB(t).QueryRow("SELECT version()").Scan(&version); err != nil {
		t.Fatalf("query DuckDB version: %v", err)
	}
	if version != "v1.4.5" {
		t.Fatalf("DuckDB version = %q, want v1.4.5", version)
	}
}

func TestSQLAnalytics(t *testing.T) {
	db := openTestDB(t)
	if _, err := db.Exec(`
		CREATE TABLE analytics(category VARCHAR, amount DECIMAL(10,2), recorded_at TIMESTAMP);
		INSERT INTO analytics VALUES
			('alpha', 10.25, '2026-01-01 10:00:00'),
			('alpha', 5.75, '2026-01-02 10:00:00'),
			('beta', 20.00, '2026-01-03 10:00:00'),
			('beta', NULL, '2026-01-04 10:00:00');
	`); err != nil {
		t.Fatalf("seed analytics data: %v", err)
	}

	rows, err := db.Query(`
		SELECT category, COUNT(*), COUNT(amount), CAST(SUM(amount) AS DOUBLE), MIN(recorded_at)
		FROM analytics GROUP BY category ORDER BY category
	`)
	if err != nil {
		t.Fatalf("run analytics query: %v", err)
	}
	defer rows.Close()

	want := []struct {
		category        string
		records, values int
		total           float64
	}{{"alpha", 2, 2, 16}, {"beta", 2, 1, 20}}

	for i, expected := range want {
		if !rows.Next() {
			t.Fatalf("missing analytics row %d", i)
		}
		var category string
		var records, values int
		var total float64
		var first time.Time
		if err := rows.Scan(&category, &records, &values, &total, &first); err != nil {
			t.Fatalf("scan analytics row: %v", err)
		}
		if category != expected.category || records != expected.records || values != expected.values || total != expected.total {
			t.Errorf("row = (%q, %d, %d, %v), want (%q, %d, %d, %v)", category, records, values, total, expected.category, expected.records, expected.values, expected.total)
		}
		if first.IsZero() {
			t.Error("timestamp result is zero")
		}
	}
	if rows.Next() {
		t.Fatal("analytics query returned unexpected extra rows")
	}
}

func TestIncludedCSVAnalytics(t *testing.T) {
	var records, campuses int
	err := openTestDB(t).QueryRow(`
		SELECT COUNT(*), COUNT(DISTINCT "Campus")
		FROM read_csv_auto('student-data.csv')
	`).Scan(&records, &campuses)
	if err != nil {
		t.Fatalf("query included CSV: %v", err)
	}
	if records != 61953 {
		t.Fatalf("record count = %d, want 61953", records)
	}
	if campuses == 0 {
		t.Fatal("expected at least one campus")
	}
}
