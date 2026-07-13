package postgres

import (
	"database/sql"
	"os"
	"testing"
)

func newTestDB(t *testing.T) (*sql.DB, func()) {
	dsn := os.Getenv("TEST_DATABASE_URI")
	if dsn == "" {
		dsn = "postgres://snippetbox:snippetbox@localhost:5432/snippetbox_test?sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}

	script, err := os.ReadFile("./testdata/setup.sql")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(string(script))
	if err != nil {
		t.Fatal(err)
	}

	return db, func() {
		script, err := os.ReadFile("./testdata/teardown.sql")
		if err != nil {
			t.Fatal(err)
		}

		_, err = db.Exec(string(script))
		if err != nil {
			t.Fatal(err)
		}

		db.Close()
	}
}
