package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
)

func Calculate(path string) (string, error) {
	if path == "" {
		return "", errors.New("path cannot be empty")
	}

	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open file %q: %w", path, err)
	}
	defer file.Close()

	hasher := sha256.New()

	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("calculate hash for %q: %w", path, err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func CalculateReader(reader io.Reader) (string, error) {
	if reader == nil {
		return "", errors.New("reader cannot be nil")
	}

	hasher := sha256.New()

	if _, err := io.Copy(hasher, reader); err != nil {
		return "", fmt.Errorf("calculate hash: %w", err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func CalculateBytes(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}