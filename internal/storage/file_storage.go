package storage

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/noedaka/go-url-shortener/internal/model"
)

// FileStorage реализует Storage интерфейс используя in-memory хранилище.
type FileStorage struct {
	filePath string
	mu       sync.RWMutex
	urls     map[string]string
}

type record struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
}

// NewPostgresStorage создает новый экземпляр FileStorage.
func NewFileStorage(filePath string) *FileStorage {
	fs := &FileStorage{
		filePath: filePath,
		urls:     make(map[string]string),
	}

	data, err := fs.loadData()
	if err == nil {
		fs.urls = data
	}

	return fs
}

// Save сохраняет сокращенный URL и оригинальный URL в хранилище указанного пользователя.
func (fs *FileStorage) Save(ctx context.Context, shortURL, originalURL, userID string) error {
	record := record{
		UUID:        uuid.New().String(),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		UserID:      userID,
	}

	if err := fs.appendRecord(record); err != nil {
		return err
	}

	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.urls[shortURL] = originalURL

	return nil
}

// Get возращает оригинальный URL по сокращенному.
func (fs *FileStorage) Get(ctx context.Context, shortURL string) (string, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	if url, exists := fs.urls[shortURL]; exists {
		return url, nil
	}
	return "", errors.New("URL not found")
}

// GetByUser возвращает все пары URL когда либо сокращенных указанным пользователем.
func (fs *FileStorage) GetByUser(ctx context.Context, userID string) ([]model.URLPair, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	records, err := fs.readAll()
	if err != nil {
		return nil, err
	}

	var urlPairs []model.URLPair
	for _, record := range records {
		if record.UserID == userID {
			pair := model.URLPair{
				ShortURL:    record.ShortURL,
				OriginalURL: record.OriginalURL,
			}
			urlPairs = append(urlPairs, pair)
		}
	}

	return urlPairs, nil
}

// DeleteByUser - заглушка, так как для in-memory не нужна.
func (fs *FileStorage) DeleteByUser(ctx context.Context, userID string, shortURL []string) error {
	return nil
}

func (fs *FileStorage) loadData() (map[string]string, error) {
	records, err := fs.readAll()
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, record := range records {
		result[record.ShortURL] = record.OriginalURL
	}

	return result, nil
}

func (fs *FileStorage) readAll() ([]record, error) {
	if _, err := os.Stat(fs.filePath); os.IsNotExist(err) {
		return []record{}, nil
	}

	file, err := os.ReadFile(fs.filePath)
	if err != nil {
		return nil, err
	}

	if len(file) == 0 {
		return []record{}, nil
	}

	var records []record
	if err := json.Unmarshal(file, &records); err != nil {
		return nil, err
	}

	return records, nil
}

func (fs *FileStorage) appendRecord(record record) error {
	// Используем отдельный мьютекс для файловых операций
	var fileMu sync.Mutex
	fileMu.Lock()
	defer fileMu.Unlock()

	file, err := os.OpenFile(fs.filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	if stat.Size() == 0 {
		if _, err := file.Write([]byte("[\n")); err != nil {
			return err
		}
	} else {
		if _, err := file.Seek(-1, io.SeekEnd); err != nil {
			return err
		}
		if _, err := file.Write([]byte(",\n")); err != nil {
			return err
		}
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(record); err != nil {
		return err
	}

	_, err = file.Write([]byte("]"))
	return err
}

func (fs *FileStorage) GetStats(ctx context.Context) (*model.Stats, error) {
	return nil, nil
}
