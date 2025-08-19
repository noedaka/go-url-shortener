package storage

import (
	"database/sql"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(db *sql.DB) (*PostgresStorage, error) {
    _, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS urls (
			id SERIAL PRIMARY KEY,
			short_url TEXT NOT NULL,
			original_url TEXT NOT NULL
        )
    `)
    if err != nil {
        return nil, err
    }

	return &PostgresStorage{db: db}, nil
}

func (ps *PostgresStorage) Save(shortURL, originalURL string) error {

	_, err := ps.db.Exec("INSERT INTO urls (short_url, original_url) VALUES ($1, $2)",
		shortURL, originalURL)

	if err != nil {
		return err
	}

	return nil
}

func (ps *PostgresStorage) Get(shortURL string) (string, error) {
	row := ps.db.QueryRow(`SELECT original_URL 
		FROM urls WHERE short_url = $1`, shortURL)

	var originalURL string
	err := row.Scan(&originalURL)
	if err != nil {
		return "", err
	}

	return originalURL, nil
}
