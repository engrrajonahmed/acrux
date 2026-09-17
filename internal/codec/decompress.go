package codec

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func Decompress(source, destination string) error {
	if source == "" {
		return errors.New("source cannot be empty")
	}

	if destination == "" {
		return errors.New("destination cannot be empty")
	}

	sourceInfo, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("stat source %q: %w", source, err)
	}

	if !sourceInfo.Mode().IsRegular() {
		return fmt.Errorf("source %q is not a regular file", source)
	}

	if filepath.Clean(source) == filepath.Clean(destination) {
		return errors.New("source and destination are the same file")
	}

	if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		return fmt.Errorf("create destination directory: %w", err)
	}

	input, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("open source %q: %w", source, err)
	}
	defer input.Close()

	reader, err := gzip.NewReader(input)
	if err != nil {
		return fmt.Errorf("open compressed source %q: %w", source, err)
	}
	defer reader.Close()

	output, err := os.OpenFile(
		destination,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0600,
	)
	if err != nil {
		return fmt.Errorf("create decompressed file %q: %w", destination, err)
	}

	success := false
	defer func() {
		if !success {
			_ = output.Close()
			_ = os.Remove(destination)
		}
	}()

	if _, err := io.Copy(output, reader); err != nil {
		return fmt.Errorf("decompress %q: %w", source, err)
	}

	if err := output.Sync(); err != nil {
		return fmt.Errorf("sync decompressed file %q: %w", destination, err)
	}

	if err := output.Close(); err != nil {
		return fmt.Errorf("close decompressed file %q: %w", destination, err)
	}

	success = true

	return nil
}

func DecompressReader(reader io.Reader, writer io.Writer) error {
	if reader == nil {
		return errors.New("reader cannot be nil")
	}

	if writer == nil {
		return errors.New("writer cannot be nil")
	}

	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return fmt.Errorf("open compressed data: %w", err)
	}
	defer gzipReader.Close()

	if _, err := io.Copy(writer, gzipReader); err != nil {
		return fmt.Errorf("decompress data: %w", err)
	}

	return nil
}