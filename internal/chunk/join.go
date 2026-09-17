package chunk

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func Join(chunks []string, destination string) error {
	if len(chunks) == 0 {
		return errors.New("at least one chunk is required")
	}

	if destination == "" {
		return errors.New("destination cannot be empty")
	}

	ordered, err := orderChunks(chunks)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		return fmt.Errorf("create destination directory: %w", err)
	}

	output, err := os.OpenFile(
		destination,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0600,
	)
	if err != nil {
		return fmt.Errorf("create destination %q: %w", destination, err)
	}

	success := false
	defer func() {
		if !success {
			_ = output.Close()
			_ = os.Remove(destination)
		}
	}()

	buffer := make([]byte, 1024*1024)

	for _, chunkPath := range ordered {
		info, err := os.Stat(chunkPath)
		if err != nil {
			return fmt.Errorf("stat chunk %q: %w", chunkPath, err)
		}

		if !info.Mode().IsRegular() {
			return fmt.Errorf("chunk %q is not a regular file", chunkPath)
		}

		input, err := os.Open(chunkPath)
		if err != nil {
			return fmt.Errorf("open chunk %q: %w", chunkPath, err)
		}

		if _, err := io.CopyBuffer(output, input, buffer); err != nil {
			_ = input.Close()
			return fmt.Errorf("join chunk %q: %w", chunkPath, err)
		}

		if err := input.Close(); err != nil {
			return fmt.Errorf("close chunk %q: %w", chunkPath, err)
		}
	}

	if err := output.Sync(); err != nil {
		return fmt.Errorf("sync joined file %q: %w", destination, err)
	}

	if err := output.Close(); err != nil {
		return fmt.Errorf("close joined file %q: %w", destination, err)
	}

	success = true

	return nil
}

func JoinFiles(chunks []string, destination string) error {
	return Join(chunks, destination)
}

func orderChunks(chunks []string) ([]string, error) {
	type numberedChunk struct {
		path   string
		number int
	}

	items := make([]numberedChunk, 0, len(chunks))

	for _, path := range chunks {
		if strings.TrimSpace(path) == "" {
			return nil, errors.New("chunk path cannot be empty")
		}

		number, err := chunkNumber(path)
		if err != nil {
			return nil, err
		}

		items = append(items, numberedChunk{
			path:   path,
			number: number,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].number < items[j].number
	})

	for i, item := range items {
		expected := i + 1
		if item.number != expected {
			return nil, fmt.Errorf(
				"missing chunk part %d",
				expected,
			)
		}
	}

	ordered := make([]string, len(items))
	for i, item := range items {
		ordered[i] = item.path
	}

	return ordered, nil
}

func chunkNumber(path string) (int, error) {
	base := filepath.Base(path)

	index := strings.LastIndex(base, ".part")
	if index == -1 {
		return 0, fmt.Errorf("invalid chunk filename %q", path)
	}

	value := base[index+len(".part"):]
	if value == "" {
		return 0, fmt.Errorf("invalid chunk number in %q", path)
	}

	number, err := strconv.Atoi(value)
	if err != nil || number <= 0 {
		return 0, fmt.Errorf("invalid chunk number in %q", path)
	}

	return number, nil
}