package file

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func Delete(path string) error {
	if path == "" {
		return errors.New("path cannot be empty")
	}

	info, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("path does not exist: %q", path)
		}

		return fmt.Errorf("stat %q: %w", path, err)
	}

	if info.IsDir() {
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("delete directory %q: %w", path, err)
		}

		return nil
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("delete file %q: %w", path, err)
	}

	return nil
}

func DeleteFile(path string) error {
	if path == "" {
		return errors.New("file path cannot be empty")
	}

	info, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("file does not exist: %q", path)
		}

		return fmt.Errorf("stat file %q: %w", path, err)
	}

	if info.IsDir() {
		return fmt.Errorf("path %q is a directory", path)
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("delete file %q: %w", path, err)
	}

	return nil
}

func DeleteDirectory(path string) error {
	if path == "" {
		return errors.New("directory path cannot be empty")
	}

	info, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("directory does not exist: %q", path)
		}

		return fmt.Errorf("stat directory %q: %w", path, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path %q is not a directory", path)
	}

	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("delete directory %q: %w", path, err)
	}

	return nil
}

func Exists(path string) (bool, error) {
	if path == "" {
		return false, errors.New("path cannot be empty")
	}

	_, err := os.Lstat(path)
	if err == nil {
		return true, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	return false, fmt.Errorf("check path %q: %w", path, err)
}

func IsDirectory(path string) (bool, error) {
	if path == "" {
		return false, errors.New("path cannot be empty")
	}

	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}

		return false, fmt.Errorf("stat %q: %w", path, err)
	}

	return info.IsDir(), nil
}

func IsFile(path string) (bool, error) {
	if path == "" {
		return false, errors.New("path cannot be empty")
	}

	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}

		return false, fmt.Errorf("stat %q: %w", path, err)
	}

	return info.Mode().IsRegular(), nil
}

func DeleteContents(path string) error {
	if path == "" {
		return errors.New("directory path cannot be empty")
	}

	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("directory does not exist: %q", path)
		}

		return fmt.Errorf("stat directory %q: %w", path, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path %q is not a directory", path)
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("read directory %q: %w", path, err)
	}

	for _, entry := range entries {
		entryPath := filepath.Join(path, entry.Name())

		if err := os.RemoveAll(entryPath); err != nil {
			return fmt.Errorf("delete %q: %w", entryPath, err)
		}
	}

	return nil
}