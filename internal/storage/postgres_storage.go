package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/noedaka/go-url-shortener/internal/model"
)

// PostgressStorage реализует Storage интерфейс используя PostgreSQL
type PostgresStorage struct {
	db *sql.DB
}

// NewPostgresStorage создает новый экземпляр PostgresStorage.
func NewPostgresStorage(db *sql.DB) (*PostgresStorage, error) {
	return &PostgresStorage{db: db}, nil
}

// Save сохраняет сокращенный URL и оригинальный URL в хранилище указанного пользователя
func (ps *PostgresStorage) Save(ctx context.Context, shortURL, originalURL, userID string) error {
	_, err := ps.db.ExecContext(ctx, "INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3)",
		shortURL, originalURL, userID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			existingShortID, err := ps.getExistingShortID(ctx, originalURL)
			if err != nil {
				return err
			}
			return model.NewUniqueViolationError(existingShortID, err)
		}
		return err
	}

	return nil
}

// Get возращает оригинальный URL по сокращенному
func (ps *PostgresStorage) Get(ctx context.Context, shortURL string) (string, error) {
	var originalURL string
	var isDeleted bool
	err := ps.db.QueryRowContext(ctx,
		"SELECT original_URL, is_deleted FROM urls WHERE short_url = $1", shortURL,
	).Scan(&originalURL, &isDeleted)

	if err != nil {
		return "", err
	}

	if isDeleted {
		return "", nil
	}

	return originalURL, nil
}

// GetByUser возвращает все пары URL когда либо сокращенных указанным пользователем
func (ps *PostgresStorage) GetByUser(ctx context.Context, userID string) ([]model.URLPair, error) {
	var urlPairs []model.URLPair
	rows, err := ps.db.QueryContext(ctx,
		"SELECT short_url, original_url FROM urls WHERE user_id = $1", userID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var urlPair model.URLPair
		err = rows.Scan(&urlPair.ShortURL, &urlPair.OriginalURL)
		if err != nil {
			return nil, err
		}

		urlPairs = append(urlPairs, urlPair)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return urlPairs, nil
}

// DeleteByUser удаляет сокращенные URL указанного пользователя
func (ps *PostgresStorage) DeleteByUser(ctx context.Context, userID string, shortURL []string) error {
	if err := ps.fanInUpdate(ctx, userID, shortURL); err != nil {
		return err
	}

	return nil
}

func (ps *PostgresStorage) GetStats(ctx context.Context) (*model.Stats, error) {
	stats := &model.Stats{}

	err := ps.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM urls").Scan(&stats.URLs)
	if err != nil {
		return nil, err
	}

	err = ps.db.QueryRowContext(ctx,
		"SELECT COUNT(DISTINCT user_id) FROM urls").Scan(&stats.Users)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (ps *PostgresStorage) updateDeletedForURLs(ctx context.Context, userID string, urls []string) error {
	tx, err := ps.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	placeholders := make([]string, len(urls))
	args := make([]interface{}, len(urls)+1)
	args[0] = userID

	for i, uri := range urls {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = uri
	}

	query := fmt.Sprintf(
		"UPDATE urls SET is_deleted = TRUE WHERE user_id = $1 AND short_url IN (%s)",
		strings.Join(placeholders, ", "),
	)

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (ps *PostgresStorage) fanInUpdate(ctx context.Context, userID string, urls []string) error {
	const batchSize = 100
	urlBatches := make([][]string, 0)

	for i := 0; i < len(urls); i += batchSize {
		end := i + batchSize
		if end > len(urls) {
			end = len(urls)
		}
		urlBatches = append(urlBatches, urls[i:end])
	}

	errors := make(chan error, len(urlBatches))
	var wg sync.WaitGroup

	for _, batch := range urlBatches {
		wg.Add(1)
		go func(batchURIs []string) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				errors <- ctx.Err()
			default:
				if err := ps.updateDeletedForURLs(ctx, userID, batchURIs); err != nil {
					errors <- err
				}
			}
		}(batch)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		if err != nil {
			return err
		}
	}

	return nil
}

func (ps *PostgresStorage) getExistingShortID(ctx context.Context, originalURL string) (string, error) {
	var shortID string
	err := ps.db.QueryRowContext(ctx,
		"SELECT short_url FROM urls WHERE original_url = $1",
		originalURL,
	).Scan(&shortID)

	if err != nil {
		return "", err
	}

	return shortID, nil
}
