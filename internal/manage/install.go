package manage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	installDirectoryName = ".local/bin"
	executableName       = "acrux"
	configDirectoryName  = "acrux"
)

type InstallResult struct {
	ExecutablePath string
	ConfigPath     string
	AlreadyInstalled bool
}

func Install(sourceExecutable string) (*InstallResult, error) {
	if sourceExecutable == "" {
		return nil, errors.New("source executable cannot be empty")
	}

	sourceInfo, err := os.Stat(sourceExecutable)
	if err != nil {
		return nil, fmt.Errorf("stat source executable %q: %w", sourceExecutable, err)
	}

	if sourceInfo.IsDir() {
		return nil, fmt.Errorf("source executable %q is a directory", sourceExecutable)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get user home directory: %w", err)
	}

	installDir := filepath.Join(home, installDirectoryName)
	target := filepath.Join(installDir, executableName)
	configDir := filepath.Join(home, ".config", configDirectoryName)

	if filepath.Clean(sourceExecutable) == filepath.Clean(target) {
		if err := os.MkdirAll(configDir, 0700); err != nil {
			return nil, fmt.Errorf("create configuration directory: %w", err)
		}

		return &InstallResult{
			ExecutablePath: target,
			ConfigPath:     configDir,
			AlreadyInstalled: true,
		}, nil
	}

	if err := os.MkdirAll(installDir, 0755); err != nil {
		return nil, fmt.Errorf("create installation directory: %w", err)
	}

	if err := copyExecutable(sourceExecutable, target, sourceInfo.Mode().Perm()); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, fmt.Errorf("create configuration directory: %w", err)
	}

	accountPath := filepath.Join(configDir, "account.conf")
	preferencePath := filepath.Join(configDir, "preference.conf")

	if err := initializeFile(accountPath); err != nil {
		return nil, fmt.Errorf("initialize account configuration: %w", err)
	}

	if err := initializeFile(preferencePath); err != nil {
		return nil, fmt.Errorf("initialize preference configuration: %w", err)
	}

	return &InstallResult{
		ExecutablePath: target,
		ConfigPath:     configDir,
		AlreadyInstalled: false,
	}, nil
}

func IsInstalled() (bool, error) {
	path, err := InstalledExecutablePath()
	if err != nil {
		return false, err
	}

	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}

		return false, fmt.Errorf("check installed executable: %w", err)
	}

	if info.IsDir() {
		return false, nil
	}

	return true, nil
}

func InstalledExecutablePath() (string, error) {
	if runtime.GOOS == "windows" {
		return "", errors.New("acrux user installation is not supported on Windows")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get user home directory: %w", err)
	}

	return filepath.Join(home, installDirectoryName, executableName), nil
}

func ConfigurationDirectory() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get user home directory: %w", err)
	}

	return filepath.Join(home, ".config", configDirectoryName), nil
}

func IsRunningFromInstalledLocation(executablePath string) (bool, error) {
	if executablePath == "" {
		return false, errors.New("executable path cannot be empty")
	}

	installedPath, err := InstalledExecutablePath()
	if err != nil {
		return false, err
	}

	source, err := filepath.Abs(executablePath)
	if err != nil {
		return false, fmt.Errorf("resolve executable path: %w", err)
	}

	target, err := filepath.Abs(installedPath)
	if err != nil {
		return false, fmt.Errorf("resolve installed executable path: %w", err)
	}

	return filepath.Clean(source) == filepath.Clean(target), nil
}

func copyExecutable(source, destination string, mode os.FileMode) error {
	input, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("open source executable %q: %w", source, err)
	}
	defer input.Close()

	output, err := os.OpenFile(
		destination,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		mode,
	)
	if err != nil {
		return fmt.Errorf("create installed executable %q: %w", destination, err)
	}

	if _, err := input.Stat(); err != nil {
		_ = output.Close()
		_ = os.Remove(destination)
		return fmt.Errorf("stat source executable: %w", err)
	}

	buffer := make([]byte, 1024*1024)

	for {
		n, readErr := input.Read(buffer)

		if n > 0 {
			if _, writeErr := output.Write(buffer[:n]); writeErr != nil {
				_ = output.Close()
				_ = os.Remove(destination)
				return fmt.Errorf("write installed executable: %w", writeErr)
			}
		}

		if readErr != nil {
			if errors.Is(readErr, os.ErrClosed) {
				_ = output.Close()
				_ = os.Remove(destination)
				return fmt.Errorf("read source executable: %w", readErr)
			}

			if readErr.Error() == "EOF" {
				break
			}

			_ = output.Close()
			_ = os.Remove(destination)
			return fmt.Errorf("read source executable: %w", readErr)
		}
	}

	if err := output.Sync(); err != nil {
		_ = output.Close()
		_ = os.Remove(destination)
		return fmt.Errorf("sync installed executable: %w", err)
	}

	if err := output.Close(); err != nil {
		_ = os.Remove(destination)
		return fmt.Errorf("close installed executable: %w", err)
	}

	if err := os.Chmod(destination, mode|0111); err != nil {
		_ = os.Remove(destination)
		return fmt.Errorf("make installed executable executable: %w", err)
	}

	return nil
}

func initializeFile(path string) error {
	file, err := os.OpenFile(
		path,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		0600,
	)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil
		}

		return err
	}

	if err := file.Close(); err != nil {
		return err
	}

	return nil
}