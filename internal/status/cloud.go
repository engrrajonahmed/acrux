package status

import (
	"context"
	"fmt"
	"strings"

	"acrux/internal/config"

	"github.com/rclone/rclone/fs"
)

type CloudStatus struct {
	Path      string
	Exists    bool
	Available bool
	Size      int64
}

func GetCloudStatus(account config.Account, remotePath string) (CloudStatus, error) {
	ctx := context.Background()

	remoteName := strings.TrimSpace(account.Label)
	if remoteName == "" {
		return CloudStatus{}, fmt.Errorf("cloud account label is required")
	}

	remotePath = strings.Trim(remotePath, "/")

	fsPath := remoteName + ":"
	if remotePath != "" {
		fsPath += remotePath
	}

	fsrc, err := fs.NewFs(ctx, fsPath)
	if err != nil {
		return CloudStatus{
			Path:      remotePath,
			Exists:    false,
			Available: false,
		}, fmt.Errorf("open cloud remote %q: %w", fsPath, err)
	}

	status := CloudStatus{
		Path:      remotePath,
		Available: true,
	}

	if remotePath == "" {
		status.Exists = true
		return status, nil
	}

	entry, err := fsrc.NewObject(ctx, remotePath)
	if err == nil {
		status.Exists = true
		status.Size = entry.Size()
		return status, nil
	}

	entries, listErr := fsrc.List(ctx, remotePath)
	if listErr == nil {
		status.Exists = true
		status.Size = 0

		for _, entry := range entries {
			if entry == nil {
				continue
			}

			if object, ok := entry.(fs.Object); ok {
				status.Size += object.Size()
			}
		}

		return status, nil
	}

	parent := remotePath
	if index := strings.LastIndex(parent, "/"); index >= 0 {
		parent = parent[:index]
	} else {
		parent = ""
	}

	entries, listErr = fsrc.List(ctx, parent)
	if listErr != nil {
		return status, nil
	}

	for _, entry := range entries {
		if entry == nil {
			continue
		}

		if entry.Remote() == remotePath {
			status.Exists = true

			if object, ok := entry.(fs.Object); ok {
				status.Size = object.Size()
			}

			return status, nil
		}
	}

	return status, nil
}

func CloudExists(account config.Account, remotePath string) (bool, error) {
	status, err := GetCloudStatus(account, remotePath)
	if err != nil {
		return false, err
	}

	return status.Exists, nil
}

func CloudSize(account config.Account, remotePath string) (int64, error) {
	status, err := GetCloudStatus(account, remotePath)
	if err != nil {
		return 0, err
	}

	if !status.Exists {
		return 0, fmt.Errorf("cloud path does not exist: %s", remotePath)
	}

	return status.Size, nil
}

func CloudAvailable(account config.Account) bool {
	ctx := context.Background()

	remoteName := strings.TrimSpace(account.Label)
	if remoteName == "" {
		return false
	}

	fsrc, err := fs.NewFs(ctx, remoteName+":")
	if err != nil {
		return false
	}

	return fsrc != nil
}