package packet

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func Export(source, destination string) error {
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

	destinationInfo, err := os.Stat(destination)
	if err == nil && destinationInfo.IsDir() {
		destination = filepath.Join(destination, filepath.Base(source))
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat destination %q: %w", destination, err)
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

	tempFile, err := os.CreateTemp(filepath.Dir(destination), ".acrux-export-*")
	if err != nil {
		return fmt.Errorf("create temporary export file: %w", err)
	}

	tempPath := tempFile.Name()
	success := false

	defer func() {
		_ = tempFile.Close()
		if !success {
			_ = os.Remove(tempPath)
		}
	}()

	if err := tempFile.Chmod(sourceInfo.Mode().Perm()); err != nil {
		return fmt.Errorf("set temporary export permissions: %w", err)
	}

	if _, err := io.Copy(tempFile, input); err != nil {
		return fmt.Errorf("export %q: %w", source, err)
	}

	if err := tempFile.Sync(); err != nil {
		return fmt.Errorf("sync exported file: %w", err)
	}

	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("close temporary export file: %w", err)
	}

	if err := os.Rename(tempPath, destination); err != nil {
		return fmt.Errorf("move exported file to %q: %w", destination, err)
	}

	success = true

	return nil
}

func ExportSources(sources []string, destination string) error {
	if len(sources) == 0 {
		return errors.New("at least one source is required")
	}

	if destination == "" {
		return errors.New("destination cannot be empty")
	}

	info, err := os.Stat(destination)
	if err == nil && !info.IsDir() {
		return fmt.Errorf("destination %q is not a directory", destination)
	}

	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat destination %q: %w", destination, err)
	}

	if err := os.MkdirAll(destination, 0755); err != nil {
		return fmt.Errorf("create destination directory: %w", err)
	}

	for _, source := range sources {
		if source == "" {
			return errors.New("source path cannot be empty")
		}

		sourceInfo, err := os.Stat(source)
		if err != nil {
			return fmt.Errorf("stat source %q: %w", source, err)
		}

		if sourceInfo.IsDir() {
			return fmt.Errorf(
				"source %q is a directory; use Archive before exporting directories",
				source,
			)
		}

		target := filepath.Join(destination, filepath.Base(source))

		if err := Export(source, target); err != nil {
			return fmt.Errorf("export %q: %w", source, err)
		}
	}

	return nil
}