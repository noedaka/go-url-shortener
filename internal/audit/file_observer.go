package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/noedaka/go-url-shortener/internal/model"
)

type FileObserver struct {
	filePath string
	file     *os.File
	mu       sync.Mutex
}

func NewFileObserver(filePath string) (*FileObserver, error) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &FileObserver{
		filePath: filePath,
		file:     file,
	}, nil
}

func (o *FileObserver) Notify(event model.AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	_, err = o.file.Write(append(data, '\n'))
	if err != nil {
		return fmt.Errorf("write audit event to file: %w", err)
	}

	return nil
}

func (o *FileObserver) Close() error {
	if o.file != nil {
		return o.file.Close()
	}
	return nil
}
