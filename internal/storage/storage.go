package storage

import (
	"encoding/json"
	"io"
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

	record := Record{
		UUID:        uuid.New().String(),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}

	return fs.appendRecord(record)
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

func (fs *FileStorage) appendRecord(record Record) error {
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
