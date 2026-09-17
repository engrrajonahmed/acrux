package status

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type LocalStatus struct {
	Path         string
	Exists       bool
	IsDirectory  bool
	IsFile       bool
	Size         int64
	FreeSpace    uint64
	TotalSpace   uint64
	UsedSpace    uint64
}

func GetLocalStatus(path string) (*LocalStatus, error) {
	if path == "" {
		return nil, errors.New("local path cannot be empty")
	}

	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &LocalStatus{
				Path:   path,
				Exists: false,
			}, nil
		}

		return nil, fmt.Errorf("stat local path %q: %w", path, err)
	}

	status := &LocalStatus{
		Path:        path,
		Exists:      true,
		IsDirectory: info.IsDir(),
		IsFile:      info.Mode().IsRegular(),
		Size:        info.Size(),
	}

	if info.IsDir() {
		total, free, err := diskSpace(path)
		if err == nil {
			status.TotalSpace = total
			status.FreeSpace = free

			if total >= free {
				status.UsedSpace = total - free
			}
		}
	}

	return status, nil
}

func LocalExists(path string) (bool, error) {
	if path == "" {
		return false, errors.New("local path cannot be empty")
	}

	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	return false, fmt.Errorf("check local path %q: %w", path, err)
}

func LocalSize(path string) (int64, error) {
	if path == "" {
		return 0, errors.New("local path cannot be empty")
	}

	info, err := os.Stat(path)
	if err != nil {
		return 0, fmt.Errorf("stat local path %q: %w", path, err)
	}

	if info.IsDir() {
		return directorySize(path)
	}

	return info.Size(), nil
}

func directorySize(path string) (int64, error) {
	var total int64

	err := filepath.Walk(path, func(currentPath string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("access %q: %w", currentPath, walkErr)
		}

		if info.Mode().IsRegular() {
			total += info.Size()
		}

		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("calculate directory size for %q: %w", path, err)
	}

	return total, nil
}

func diskSpace(path string) (total, free uint64, err error) {
	return 0, 0, errors.New("disk space information is not available on this platform implementation")
}