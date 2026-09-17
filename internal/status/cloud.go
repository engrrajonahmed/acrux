package status

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/config"
	"github.com/rclone/rclone/fs/operations"
)

type CloudStatus struct {
	Remote      string
	Path        string
	Exists      bool
	IsDirectory bool
	IsFile      bool
	Size        int64
	Modified    time.Time
}

func GetCloudStatus(remote, remotePath string) (*CloudStatus, error) {
	if strings.TrimSpace(remote) == "" {
		return nil, errors.New("cloud remote cannot be empty")
	}

	remotePath = strings.TrimPrefix(strings.TrimSpace(remotePath), "/")

	ctx := context.Background()

	fsrc, err := config.NewFs(ctx, remote+":"+remotePath)
	if err != nil {
		return nil, fmt.Errorf("initialize cloud remote %q: %w", remote, err)
	}

	status := &CloudStatus{
		Remote: remote,
		Path:   remotePath,
	}

	if remotePath == "" {
		status.Exists = true
		status.IsDirectory = true
		return status, nil
	}

	parentPath := remotePath
	name := remotePath

	if index := strings.LastIndex(remotePath, "/"); index >= 0 {
		parentPath = remotePath[:index]
		name = remotePath[index+1:]
	}

	parentFS, err := config.NewFs(ctx, remote+":"+parentPath)
	if err != nil {
		return nil, fmt.Errorf("initialize cloud parent %q: %w", parentPath, err)
	}

	entries, err := operations.List(ctx, parentFS, "")
	if err != nil {
		return nil, fmt.Errorf("list cloud path %q: %w", parentPath, err)
	}

	for _, entry := range entries {
		if strings.TrimSuffix(entry.Remote(), "/") != name {
			continue
		}

		status.Exists = true

		switch object := entry.(type) {
		case fs.Directory:
			status.IsDirectory = true
			status.Size = object.Size()
			status.Modified = object.ModTime(ctx)

		case fs.Object:
			status.IsFile = true
			status.Size = object.Size()
			status.Modified = object.ModTime(ctx)

		default:
			return nil, fmt.Errorf("unsupported cloud item type for %q", remotePath)
		}

		return status, nil
	}

	return status, nil
}

func CloudExists(remote, remotePath string) (bool, error) {
	status, err := GetCloudStatus(remote, remotePath)
	if err != nil {
		return false, err
	}

	return status.Exists, nil
}

func CloudSize(remote, remotePath string) (int64, error) {
	status, err := GetCloudStatus(remote, remotePath)
	if err != nil {
		return 0, err
	}

	if !status.Exists {
		return 0, fmt.Errorf("cloud path does not exist: %s:%s", remote, remotePath)
	}

	return status.Size, nil
}

func CloudAvailable(remote string) (bool, error) {
	if strings.TrimSpace(remote) == "" {
		return false, errors.New("cloud remote cannot be empty")
	}

	ctx := context.Background()

	fsrc, err := config.NewFs(ctx, remote+":")
	if err != nil {
		return false, fmt.Errorf("initialize cloud remote %q: %w", remote, err)
	}

	_, err = operations.List(ctx, fsrc, "")
	if err != nil {
		return false, fmt.Errorf("access cloud remote %q: %w", remote, err)
	}

	return true, nil
}