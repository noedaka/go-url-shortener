package dbc

import "database/sql"

func InitDatabase(db *sql.DB) error {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS urls (
			id SERIAL PRIMARY KEY,
			short_url TEXT NOT NULL,
			original_url TEXT NOT NULL,
			user_id TEXT NOT NULL
        )
    `)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_og_url
		ON urls (original_url)
	`)
	if err != nil {
		return err
	}

	return nil
}
