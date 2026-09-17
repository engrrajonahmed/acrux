package packet

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func Archive(sources []string, destination string) error {
	if len(sources) == 0 {
		return errors.New("at least one source is required")
	}

	if destination == "" {
		return errors.New("destination cannot be empty")
	}

	if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		return fmt.Errorf("create archive directory: %w", err)
	}

	output, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("create archive %q: %w", destination, err)
	}
	defer output.Close()

	tw := tar.NewWriter(output)
	defer tw.Close()

	for _, source := range sources {
		if source == "" {
			return errors.New("source path cannot be empty")
		}

		info, err := os.Stat(source)
		if err != nil {
			return fmt.Errorf("stat source %q: %w", source, err)
		}

		base := filepath.Base(filepath.Clean(source))

		err = filepath.Walk(source, func(path string, fileInfo os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return fmt.Errorf("access %q: %w", path, walkErr)
			}

			relative, err := filepath.Rel(source, path)
			if err != nil {
				return fmt.Errorf("calculate relative path for %q: %w", path, err)
			}

			archivePath := base
			if relative != "." {
				archivePath = filepath.Join(base, relative)
			}

			header, err := tar.FileInfoHeader(fileInfo, "")
			if err != nil {
				return fmt.Errorf("create archive header for %q: %w", path, err)
			}

			header.Name = filepath.ToSlash(archivePath)

			if err := tw.WriteHeader(header); err != nil {
				return fmt.Errorf("write archive header for %q: %w", path, err)
			}

			if !fileInfo.Mode().IsRegular() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("open source file %q: %w", path, err)
			}

			if _, err := io.Copy(tw, file); err != nil {
				_ = file.Close()
				return fmt.Errorf("write %q to archive: %w", path, err)
			}

			if err := file.Close(); err != nil {
				return fmt.Errorf("close source file %q: %w", path, err)
			}

			return nil
		})

		if err != nil {
			return fmt.Errorf("archive source %q: %w", source, err)
		}

		_ = info
	}

	if err := tw.Close(); err != nil {
		return fmt.Errorf("finalize archive: %w", err)
	}

	if err := output.Close(); err != nil {
		return fmt.Errorf("close archive %q: %w", destination, err)
	}

	return nil
}