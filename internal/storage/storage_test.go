package storage

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testFilePath = "test_storage.json"

func cleanup() {
	_ = os.Remove(testFilePath)
}

func TestNewFileStorage(t *testing.T) {
	defer cleanup()
	fs := NewFileStorage(testFilePath)
	assert.NotNil(t, fs, "Expected FileStorage instance, got nil")
}

func TestSaveAndGet(t *testing.T) {
	defer cleanup()
	ctx := context.Background()
	fs := NewFileStorage(testFilePath)

	err := fs.Save(ctx, "abc", "https://example.com", "")
	assert.NoError(t, err, "Save failed")

	url, err := fs.Get(ctx, "abc")
	assert.NoError(t, err, "Get failed")
	assert.Equal(t, "https://example.com", url, "URL mismatch")
}

func TestGetNonExistent(t *testing.T) {
	defer cleanup()
	ctx := context.Background()
	fs := NewFileStorage(testFilePath)

	_, err := fs.Get(ctx, "nonexistent")
	assert.Error(t, err, "Expected error for non-existent key")
}

func TestMultipleSaves(t *testing.T) {
	defer cleanup()
	ctx := context.Background()
	fs := NewFileStorage(testFilePath)

	assert.NoError(t, fs.Save(ctx, "k1", "v1", ""), "Save k1 failed")
	assert.NoError(t, fs.Save(ctx, "k2", "v2", ""), "Save k2 failed")

	val1, err1 := fs.Get(ctx, "k1")
	val2, err2 := fs.Get(ctx, "k2")

	assert.NoError(t, err1, "Get k1 failed")
	assert.NoError(t, err2, "Get k2 failed")
	assert.Equal(t, "v1", val1, "Value for k1 mismatch")
	assert.Equal(t, "v2", val2, "Value for k2 mismatch")
}

func TestSaveEmptyValues(t *testing.T) {
	defer cleanup()
	ctx := context.Background()
	fs := NewFileStorage(testFilePath)

	err := fs.Save(ctx, "", "", "")
	assert.NoError(t, err, "Save with empty values failed")

	val, err := fs.Get(ctx, "")
	assert.NoError(t, err, "Get for empty key failed")
	assert.Equal(t, "", val, "Expected empty string for empty key")
}

func BenchmarkSave(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		cleanup()
		fs := NewFileStorage(testFilePath)
		b.StartTimer()

		err := fs.Save(ctx, "test-key", "https://example.com", "")
		if err != nil {
			b.Fatalf("Save failed: %v", err)
		}
	}
}

func BenchmarkGet(b *testing.B) {
	ctx := context.Background()
	fs := NewFileStorage(testFilePath)

	err := fs.Save(ctx, "test-key", "https://example.com", "")
	if err != nil {
		b.Fatalf("Setup failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := fs.Get(ctx, "test-key")
		if err != nil {
			b.Fatalf("Get failed: %v", err)
		}
	}
}

func BenchmarkSaveAndGet(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		cleanup()
		fs := NewFileStorage(testFilePath)
		b.StartTimer()

		key := "test-key"
		value := "https://example.com"
		err := fs.Save(ctx, key, value, "")
		if err != nil {
			b.Fatalf("Save failed: %v", err)
		}

		_, err = fs.Get(ctx, key)
		if err != nil {
			b.Fatalf("Get failed: %v", err)
		}
	}
}

func BenchmarkMultipleSaves(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		cleanup()
		fs := NewFileStorage(testFilePath)
		b.StartTimer()

		for j := 0; j < 10; j++ {
			key := string(rune('a' + j))
			value := "https://example.com/" + key
			err := fs.Save(ctx, key, value, "")
			if err != nil {
				b.Fatalf("Save failed for key %s: %v", key, err)
			}
		}
	}
}
