package file

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func Copy(source, destination string) error {
	if source == "" {
		return errors.New("source path cannot be empty")
	}

	if destination == "" {
		return errors.New("destination path cannot be empty")
	}

	sourceInfo, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("stat source %q: %w", source, err)
	}

	destinationInfo, err := os.Stat(destination)
	if err == nil && destinationInfo.Mode().IsRegular() && sourceInfo.IsDir() {
		return fmt.Errorf("destination %q is a file", destination)
	}

	if sourceInfo.IsDir() {
		return CopyDirectory(source, destination)
	}

	if !sourceInfo.Mode().IsRegular() {
		return fmt.Errorf("unsupported source type %q", source)
	}

	return CopyFile(source, destination)
}

func CopyFile(source, destination string) error {
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

	output, err := os.OpenFile(
		destination,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		sourceInfo.Mode().Perm(),
	)
	if err != nil {
		return fmt.Errorf("create destination %q: %w", destination, err)
	}

	if _, err := io.Copy(output, input); err != nil {
		_ = output.Close()
		_ = os.Remove(destination)
		return fmt.Errorf("copy %q to %q: %w", source, destination, err)
	}

	if err := output.Close(); err != nil {
		_ = os.Remove(destination)
		return fmt.Errorf("close destination %q: %w", destination, err)
	}

	if err := os.Chmod(destination, sourceInfo.Mode().Perm()); err != nil {
		return fmt.Errorf("preserve permissions for %q: %w", destination, err)
	}

	return nil
}

func CopyDirectory(source, destination string) error {
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("stat source directory %q: %w", source, err)
	}

	if !sourceInfo.IsDir() {
		return fmt.Errorf("source %q is not a directory", source)
	}

	destinationInfo, err := os.Stat(destination)
	if err == nil {
		if !destinationInfo.IsDir() {
			return fmt.Errorf("destination %q is not a directory", destination)
		}

		destination = filepath.Join(destination, filepath.Base(filepath.Clean(source)))
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat destination %q: %w", destination, err)
	}

	if filepath.Clean(source) == filepath.Clean(destination) {
		return errors.New("source and destination are the same directory")
	}

	if err := os.MkdirAll(destination, sourceInfo.Mode().Perm()); err != nil {
		return fmt.Errorf("create destination directory %q: %w", destination, err)
	}

	err = filepath.Walk(source, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("access %q: %w", path, walkErr)
		}

		relativePath, err := filepath.Rel(source, path)
		if err != nil {
			return fmt.Errorf("calculate relative path for %q: %w", path, err)
		}

		if relativePath == "." {
			return nil
		}

		targetPath := filepath.Join(destination, relativePath)

		if info.IsDir() {
			if err := os.MkdirAll(targetPath, info.Mode().Perm()); err != nil {
				return fmt.Errorf("create directory %q: %w", targetPath, err)
			}

			return nil
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		if err := CopyFile(path, targetPath); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("copy directory %q to %q: %w", source, destination, err)
	}

	return nil
}