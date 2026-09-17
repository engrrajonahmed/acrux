package browse

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"acrux/internal/config"

	"github.com/rclone/rclone/fs"
)

type CloudItem struct {
	Name       string
	Path       string
	IsDir      bool
	Size       int64
	ModifiedAt time.Time
	CreatedAt  time.Time
	ID         string
}

func ListCloud(account config.Account, remotePath string) ([]CloudItem, error) {
	ctx := context.Background()

	remoteName := strings.TrimSpace(account.Label)
	if remoteName == "" {
		return nil, fmt.Errorf("cloud account label is required")
	}

	remotePath = strings.Trim(remotePath, "/")

	fsPath := remoteName + ":"
	if remotePath != "" {
		fsPath += remotePath
	}

	fsrc, err := fs.NewFs(ctx, fsPath)
	if err != nil {
		return nil, fmt.Errorf("open cloud remote %q: %w", fsPath, err)
	}

	entries, err := fsrc.List(ctx, remotePath)
	if err != nil {
		return nil, fmt.Errorf("list cloud path %q: %w", remotePath, err)
	}

	items := make([]CloudItem, 0, len(entries))

	for _, entry := range entries {
		if entry == nil {
			continue
		}

		_, isDir := entry.(fs.Directory)

		item := CloudItem{
			Name:       path.Base(entry.Remote()),
			Path:       entry.Remote(),
			IsDir:      isDir,
			ModifiedAt: entry.ModTime(ctx),
		}

		if object, ok := entry.(fs.Object); ok {
			item.Size = object.Size()
		}

		if item.Name == "." || item.Name == "/" || item.Name == "" {
			item.Name = entry.Remote()
		}

		if idEntry, ok := entry.(fs.IDer); ok {
			item.ID = idEntry.ID()
		}

		items = append(items, item)
	}

	return items, nil
}

func BrowseCloud(account config.Account, remotePath string) ([]CloudItem, error) {
	return ListCloud(account, remotePath)
}

func CloudItemInfo(account config.Account, remotePath string) (CloudItem, error) {
	ctx := context.Background()

	remoteName := strings.TrimSpace(account.Label)
	if remoteName == "" {
		return CloudItem{}, fmt.Errorf("cloud account label is required")
	}

	remotePath = strings.Trim(remotePath, "/")

	if remotePath == "" {
		return CloudItem{
			Name:  remoteName,
			Path:  "",
			IsDir: true,
		}, nil
	}

	fsPath := remoteName + ":" + remotePath

	fsrc, err := fs.NewFs(ctx, fsPath)
	if err != nil {
		return CloudItem{}, fmt.Errorf("open cloud remote %q: %w", fsPath, err)
	}

	entry, err := fsrc.NewObject(ctx, remotePath)
	if err == nil {
		item := CloudItem{
			Name:       path.Base(remotePath),
			Path:       remotePath,
			IsDir:      false,
			Size:       entry.Size(),
			ModifiedAt: entry.ModTime(ctx),
		}

		if idEntry, ok := entry.(fs.IDer); ok {
			item.ID = idEntry.ID()
		}

		return item, nil
	}

	parent := path.Dir(remotePath)
	base := path.Base(remotePath)

	entries, listErr := fsrc.List(ctx, parent)
	if listErr != nil {
		return CloudItem{}, fmt.Errorf("inspect cloud path %q: %w", remotePath, listErr)
	}

	for _, candidate := range entries {
		if candidate == nil || candidate.Remote() != remotePath {
			continue
		}

		_, isDir := candidate.(fs.Directory)

		item := CloudItem{
			Name:       base,
			Path:       candidate.Remote(),
			IsDir:      isDir,
			ModifiedAt: candidate.ModTime(ctx),
		}

		if object, ok := candidate.(fs.Object); ok {
			item.Size = object.Size()
		}

		if idEntry, ok := candidate.(fs.IDer); ok {
			item.ID = idEntry.ID()
		}

		return item, nil
	}

	return CloudItem{}, fmt.Errorf("cloud item not found: %s", remotePath)
}

func CloudParent(remotePath string) string {
	remotePath = strings.Trim(remotePath, "/")
	if remotePath == "" {
		return ""
	}

	parent := path.Dir(remotePath)
	if parent == "." {
		return ""
	}

	return strings.Trim(parent, "/")
}