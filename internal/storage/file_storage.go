package storage

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"

	"github.com/google/uuid"
)

type FileStorage struct {
	filePath string
	mu       sync.Mutex
	urls     map[string]string
}

type record struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewFileStorage(filePath string) *FileStorage {
	fs := &FileStorage{
		filePath: filePath,
		urls:     make(map[string]string),
	}

	if data, err := fs.load(); err == nil {
		fs.urls = data
	}

	return fs
}

func (fs *FileStorage) Save(shortURL, originalURL string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	record := record{
		UUID:        uuid.New().String(),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}

	fs.urls[shortURL] = originalURL

	return fs.appendRecord(record)
}

func (fs *FileStorage) Get(shortURL string) (string, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if url, exists := fs.urls[shortURL]; exists {
		return url, nil
	}
	return "", errors.New("URL not found")
}

func (fs *FileStorage) load() (map[string]string, error) {
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

func (fs *FileStorage) readAll() ([]record, error) {
	if _, err := os.Stat(fs.filePath); os.IsNotExist(err) {
		return []record{}, nil
	}

	file, err := os.ReadFile(fs.filePath)
	if err != nil {
		return nil, err
	}

	var records []record
	if len(file) == 0 {
		return records, nil
	}

	if err := json.Unmarshal(file, &records); err != nil {
		return nil, err
	}

	return records, nil
}

func (fs *FileStorage) appendRecord(record record) error {
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
