package storage

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/google/uuid"
)

type URLStorage interface {
	Save(shortURL, longURL string) error
	Load() (map[string]string, error)
}

type Record struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileStorage struct {
	filePath string
	mu       sync.Mutex
}

func NewFileStorage(filePath string) *FileStorage {
	return &FileStorage{
		filePath: filePath,
	}
}

func (fs *FileStorage) Save(shortURL, originalURL string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	records, err := fs.readAll()
	if err != nil {
		return err
	}

	records = append(records, Record{
		UUID:        uuid.New().String(),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	})

	return fs.writeAll(records)
}

func (fs *FileStorage) Load() (map[string]string, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

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

func (fs *FileStorage) readAll() ([]Record, error) {
	if _, err := os.Stat(fs.filePath); os.IsNotExist(err) {
		return []Record{}, nil
	}

	file, err := os.ReadFile(fs.filePath)
	if err != nil {
		return nil, err
	}

	var records []Record
	if len(file) == 0 {
		return records, nil
	}

	if err := json.Unmarshal(file, &records); err != nil {
		return nil, err
	}

	return records, nil
}

func (fs *FileStorage) writeAll(records []Record) error {
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(fs.filePath, data, 0644)
}
