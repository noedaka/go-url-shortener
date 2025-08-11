package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testFilePath = "test_storage.json"

func cleanup() {
	_ = os.Remove(testFilePath)
}

func TestNewFileStorage(t *testing.T) {
	fs := NewFileStorage(testFilePath)
	if fs == nil {
		t.Error("Expected FileStorage instance, got nil")
	}
	defer cleanup()
}

func TestSaveAndLoad(t *testing.T) {
	fs := NewFileStorage(testFilePath)
	defer cleanup()

	err := fs.Save("abc", "https://example.com")
	if err != nil {
		t.Errorf("Save failed: %v", err)
	}

	data, err := fs.Load()
	if err != nil {
		t.Errorf("Load failed: %v", err)
	}

	if data["abc"] != "https://example.com" {
		t.Errorf("Expected https://example.com, got %s", data["abc"])
	}
}

func TestLoadEmpty(t *testing.T) {
	fs := NewFileStorage(testFilePath)
	defer cleanup()

	data, err := fs.Load()
	if err != nil {
		t.Errorf("Load failed: %v", err)
	}

	if len(data) != 0 {
		t.Errorf("Expected empty map, got %v", data)
	}
}

func TestMultipleSaves(t *testing.T) {
	fs := NewFileStorage(testFilePath)
	defer cleanup()

	assert.NoError(t, fs.Save("k1", "v1"), "Save k1 failed")
	assert.NoError(t, fs.Save("k2", "v2"), "Save k2 failed")

	data, err := fs.Load()
	assert.NoError(t, err, "Load failed")
	assert.Equal(t, 2, len(data), "Expected 2 items")
}

func TestSaveEmptyValues(t *testing.T) {
	fs := NewFileStorage(testFilePath)
	defer cleanup()

	err := fs.Save("", "")
	if err != nil {
		t.Errorf("Save with empty values failed: %v", err)
	}
}
