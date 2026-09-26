package filestore

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type LocalStore struct{ root string }

func NewLocalStore(root string) (*LocalStore, error) {
	if strings.TrimSpace(root) == "" {
		return nil, errors.New("upload directory is required")
	}
	if err := os.MkdirAll(root, 0o750); err != nil {
		return nil, fmt.Errorf("create upload directory: %w", err)
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve upload directory: %w", err)
	}
	return &LocalStore{root: absolute}, nil
}

func (s *LocalStore) Save(_ context.Context, _ string, mimeType string, src io.Reader) (SavedFile, error) {
	extension, ok := map[string]string{"application/pdf": ".pdf", "application/vnd.openxmlformats-officedocument.wordprocessingml.document": ".docx", "text/plain": ".txt"}[mimeType]
	if !ok {
		return SavedFile{}, fmt.Errorf("unsupported MIME type %q", mimeType)
	}
	key := uuid.NewString() + extension
	temporary, err := os.CreateTemp(s.root, ".upload-*")
	if err != nil {
		return SavedFile{}, err
	}
	temporaryName := temporary.Name()
	defer func() { _ = temporary.Close(); _ = os.Remove(temporaryName) }()
	hash := sha256.New()
	size, err := io.Copy(io.MultiWriter(temporary, hash), src)
	if err != nil {
		return SavedFile{}, err
	}
	if err = temporary.Sync(); err != nil {
		return SavedFile{}, err
	}
	if err = temporary.Close(); err != nil {
		return SavedFile{}, err
	}
	if err = os.Rename(temporaryName, filepath.Join(s.root, key)); err != nil {
		return SavedFile{}, err
	}
	return SavedFile{Key: key, Size: size, SHA256: hex.EncodeToString(hash.Sum(nil))}, nil
}
func (s *LocalStore) Open(_ context.Context, key string) (io.ReadCloser, error) {
	path, err := s.safePath(key)
	if err != nil {
		return nil, err
	}
	return os.Open(path)
}
func (s *LocalStore) Delete(_ context.Context, key string) error {
	path, err := s.safePath(key)
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
func (s *LocalStore) safePath(key string) (string, error) {
	if key == "" || filepath.Base(key) != key || strings.ContainsAny(key, "/\\") {
		return "", errors.New("invalid storage key")
	}
	return filepath.Join(s.root, key), nil
}
