package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	configDirectoryName = "acrux"
	preferenceFileName  = "preference.conf"

	defaultMaxArchiveSize int64 = 1024 * 1024 * 1024 // 1 GB
)

type Preference struct {
	StartingLocalDirectory string
	EncryptionKey          string
	MaxArchiveSize         int64
}

func DefaultPreference() (*Preference, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get user home directory: %w", err)
	}

	return &Preference{
		StartingLocalDirectory: home,
		EncryptionKey:          "",
		MaxArchiveSize:         defaultMaxArchiveSize,
	}, nil
}

func PrefConfigDirectory() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get user home directory: %w", err)
	}

	return filepath.Join(home, ".config", configDirectoryName), nil
}

func PreferenceFilePath() (string, error) {
	configDir, err := PrefConfigDirectory()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, preferenceFileName), nil
}

func LoadPreference() (*Preference, error) {
	path, err := PreferenceFilePath()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DefaultPreference()
		}

		return nil, fmt.Errorf("open preference file: %w", err)
	}
	defer file.Close()

	preference, err := DefaultPreference()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		switch key {
		case "starting_local_directory":
			if value != "" {
				preference.StartingLocalDirectory = value
			}

		case "encryption_key":
			preference.EncryptionKey = value

		case "max_archive_size":
			size, err := strconv.ParseInt(value, 10, 64)
			if err != nil || size <= 0 {
				continue
			}

			preference.MaxArchiveSize = size
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read preference file: %w", err)
	}

	return preference, nil
}

func SavePreference(preference *Preference) error {
	if preference == nil {
		return errors.New("preference cannot be nil")
	}

	if preference.MaxArchiveSize <= 0 {
		return errors.New("maximum archive size must be greater than zero")
	}

	if preference.StartingLocalDirectory == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("get user home directory: %w", err)
		}

		preference.StartingLocalDirectory = home
	}

	configDir, err := PrefConfigDirectory()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("create configuration directory: %w", err)
	}

	path := filepath.Join(configDir, preferenceFileName)

	content := fmt.Sprintf(
		"# acrux application preferences\n"+
			"starting_local_directory=%s\n"+
			"encryption_key=%s\n"+
			"max_archive_size=%d\n",
		preference.StartingLocalDirectory,
		preference.EncryptionKey,
		preference.MaxArchiveSize,
	)

	file, err := os.OpenFile(
		path,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0600,
	)
	if err != nil {
		return fmt.Errorf("open preference file for writing: %w", err)
	}

	if _, err := file.WriteString(content); err != nil {
		_ = file.Close()
		return fmt.Errorf("write preference file: %w", err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("close preference file: %w", err)
	}

	return nil
}

func InitializePreference() error {
	preference, err := DefaultPreference()
	if err != nil {
		return err
	}

	return SavePreference(preference)
}

func ResetPreference() error {
	path, err := PreferenceFilePath()
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return fmt.Errorf("remove preference file: %w", err)
	}

	return nil
}