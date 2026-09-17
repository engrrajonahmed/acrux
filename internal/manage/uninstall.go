package manage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func Uninstall() error {
	installedPath, err := InstalledExecutablePath()
	if err != nil {
		return err
	}

	configDir, err := ConfigurationDirectory()
	if err != nil {
		return err
	}

	if err := removeInstalledExecutable(installedPath); err != nil {
		return err
	}

	if err := removeConfigurationDirectory(configDir); err != nil {
		return err
	}

	return nil
}

func removeInstalledExecutable(path string) error {
	if path == "" {
		return errors.New("installed executable path cannot be empty")
	}

	if err := os.Remove(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return fmt.Errorf("remove installed executable %q: %w", path, err)
	}

	return nil
}

func removeConfigurationDirectory(path string) error {
	if path == "" {
		return errors.New("configuration directory cannot be empty")
	}

	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("remove configuration directory %q: %w", path, err)
	}

	return nil
}

func UninstallPaths() (executablePath, configPath string, err error) {
	executablePath, err = InstalledExecutablePath()
	if err != nil {
		return "", "", err
	}

	configPath, err = ConfigurationDirectory()
	if err != nil {
		return "", "", err
	}

	return filepath.Clean(executablePath), filepath.Clean(configPath), nil
}