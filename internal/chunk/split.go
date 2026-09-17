package chunk

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

func Split(source, destinationDir string, maxSize int64) ([]string, error) {
	if source == "" {
		return nil, errors.New("source cannot be empty")
	}

	if destinationDir == "" {
		return nil, errors.New("destination directory cannot be empty")
	}

	if maxSize <= 0 {
		return nil, errors.New("maximum chunk size must be greater than zero")
	}

	info, err := os.Stat(source)
	if err != nil {
		return nil, fmt.Errorf("stat source %q: %w", source, err)
	}

	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("source %q is not a regular file", source)
	}

	if err := os.MkdirAll(destinationDir, 0755); err != nil {
		return nil, fmt.Errorf("create destination directory: %w", err)
	}

	input, err := os.Open(source)
	if err != nil {
		return nil, fmt.Errorf("open source %q: %w", source, err)
	}
	defer input.Close()

	baseName := filepath.Base(source)
	var chunks []string
	var partNumber int

	buffer := make([]byte, 1024*1024)

	for {
		partNumber++

		chunkPath := filepath.Join(
			destinationDir,
			baseName+".part"+strconv.Itoa(partNumber),
		)

		output, err := os.OpenFile(
			chunkPath,
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
			0600,
		)
		if err != nil {
			removeChunks(chunks)
			return nil, fmt.Errorf("create chunk %q: %w", chunkPath, err)
		}

		var written int64

		for written < maxSize {
			remaining := maxSize - written
			readSize := int64(len(buffer))

			if remaining < readSize {
				readSize = remaining
			}

			n, readErr := input.Read(buffer[:readSize])

			if n > 0 {
				if _, err := output.Write(buffer[:n]); err != nil {
					_ = output.Close()
					_ = os.Remove(chunkPath)
					removeChunks(chunks)
					return nil, fmt.Errorf("write chunk %q: %w", chunkPath, err)
				}

				written += int64(n)
			}

			if readErr != nil {
				if errors.Is(readErr, io.EOF) {
					break
				}

				_ = output.Close()
				_ = os.Remove(chunkPath)
				removeChunks(chunks)
				return nil, fmt.Errorf("read source %q: %w", source, readErr)
			}
		}

		if err := output.Sync(); err != nil {
			_ = output.Close()
			_ = os.Remove(chunkPath)
			removeChunks(chunks)
			return nil, fmt.Errorf("sync chunk %q: %w", chunkPath, err)
		}

		if err := output.Close(); err != nil {
			_ = os.Remove(chunkPath)
			removeChunks(chunks)
			return nil, fmt.Errorf("close chunk %q: %w", chunkPath, err)
		}

		chunks = append(chunks, chunkPath)

		if written < maxSize {
			break
		}
	}

	return chunks, nil
}

func SplitFile(source, destinationDir string, maxSize int64) ([]string, error) {
	return Split(source, destinationDir, maxSize)
}

func removeChunks(chunks []string) {
	for _, path := range chunks {
		_ = os.Remove(path)
	}
}