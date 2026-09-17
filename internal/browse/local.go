package browse

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type LocalItem struct {
	Name    string
	Path    string
	IsDir   bool
	Size    int64
	Created time.Time
	Modified time.Time
}

func ListLocal(path string) ([]LocalItem, error) {
	if path == "" {
		return nil, errors.New("local path cannot be empty")
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat local path %q: %w", path, err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("local path %q is not a directory", path)
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("read local directory %q: %w", path, err)
	}

	items := make([]LocalItem, 0, len(entries))

	for _, entry := range entries {
		entryPath := filepath.Join(path, entry.Name())

		entryInfo, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("stat local item %q: %w", entryPath, err)
		}

		item := LocalItem{
			Name:    entry.Name(),
			Path:    entryPath,
			IsDir:   entryInfo.IsDir(),
			Size:    entryInfo.Size(),
			Modified: entryInfo.ModTime(),
		}

		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].IsDir != items[j].IsDir {
			return items[i].IsDir
		}

		return items[i].Name < items[j].Name
	})

	return items, nil
}

func BrowseLocal(path string) ([]LocalItem, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("get user home directory: %w", err)
		}

		path = home
	}

	return ListLocal(path)
}

func LocalItemInfo(path string) (*LocalItem, error) {
	if path == "" {
		return nil, errors.New("local path cannot be empty")
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat local item %q: %w", path, err)
	}

	return &LocalItem{
		Name:     filepath.Base(path),
		Path:     path,
		IsDir:    info.IsDir(),
		Size:     info.Size(),
		Modified: info.ModTime(),
	}, nil
}

func LocalParent(path string) (string, error) {
	if path == "" {
		return "", errors.New("local path cannot be empty")
	}

	cleanPath := filepath.Clean(path)
	parent := filepath.Dir(cleanPath)

	if parent == cleanPath {
		return cleanPath, nil
	}

	return parent, nil
}