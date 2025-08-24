package storage

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/noedaka/go-url-shortener/internal/model"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(db *sql.DB) (*PostgresStorage, error) {
	return &PostgresStorage{db: db}, nil
}

func (ps *PostgresStorage) Save(shortURL, originalURL string) error {
	_, err := ps.db.Exec("INSERT INTO urls (short_url, original_url) VALUES ($1, $2)",
		shortURL, originalURL)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			existingShortID, err := ps.getExistingShortID(originalURL)
			if err != nil {
				return err
			}
			return model.NewUniqueViolationError(existingShortID, err)
		}
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

func (ps *PostgresStorage) getExistingShortID(originalURL string) (string, error) {
	var shortID string
	err := ps.db.QueryRow(
		"SELECT short_url FROM urls WHERE original_url = $1",
		originalURL,
	).Scan(&shortID)

	if err != nil {
		return "", err
	}

	return shortID, nil
}
