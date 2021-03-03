package cache

import (
	"database/sql"
)

func CreateOrUpdateDatabase(db *sql.DB) error {
	return createRegionsTable(db)
}

func createTable(db *sql.DB, ct string) error {
	stmt, err := db.Prepare(ct)
	if err != nil {
		return err
	}
	_, err = stmt.Exec()
	return err
}

func createRegionsTable(db *sql.DB) error {
	createRegionsTableSQL := `CREATE TABLE IF NOT EXISTS Regions (
		"ID" TEXT NOT NULL PRIMARY KEY,
		"DisplayName" TEXT NOT NULL,
		"Country" TEXT,
		"Continent" TEXT,
		"Special" INTEGER
	);`
	if err := createTable(db, createRegionsTableSQL); err != nil {
		return err
	}
	return nil
}
