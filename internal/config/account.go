package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const accountFileName = "account.conf"

type Account struct {
	Label      string
	ClientID   string
	SecretID   string
	RootFolder string
	RemoteType string
	Values     map[string]string
}

func AccConfigDirectory() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get user home directory: %w", err)
	}

	return filepath.Join(home, ".config", ".acrux"), nil
}

func AccountFilePath() (string, error) {
	configDir, err := AccConfigDirectory()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, accountFileName), nil
}

func LoadAccounts() ([]Account, error) {
	path, err := AccountFilePath()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []Account{}, nil
		}

		return nil, fmt.Errorf("open account file: %w", err)
	}
	defer file.Close()

	var accounts []Account
	var current *Account

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if line == "[account]" {
			if current != nil {
				accounts = append(accounts, *current)
			}

			current = &Account{
				Values: make(map[string]string),
			}

			continue
		}

		if current == nil {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		switch key {
		case "label":
			current.Label = value
		case "client_id":
			current.ClientID = value
		case "secret_id":
			current.SecretID = value
		case "root_folder_id":
			current.RootFolder = value
		case "remote_type":
			current.RemoteType = value
		default:
			current.Values[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read account file: %w", err)
	}

	if current != nil {
		accounts = append(accounts, *current)
	}

	return accounts, nil
}

func SaveAccounts(accounts []Account) error {
	configDir, err := AccConfigDirectory()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("create configuration directory: %w", err)
	}

	path := filepath.Join(configDir, accountFileName)

	file, err := os.OpenFile(
		path,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0600,
	)
	if err != nil {
		return fmt.Errorf("open account file for writing: %w", err)
	}

	for _, account := range accounts {
		if _, err := file.WriteString("[account]\n"); err != nil {
			_ = file.Close()
			return fmt.Errorf("write account section: %w", err)
		}

		values := []struct {
			key   string
			value string
		}{
			{"label", account.Label},
			{"client_id", account.ClientID},
			{"secret_id", account.SecretID},
			{"root_folder_id", account.RootFolder},
			{"remote_type", account.RemoteType},
		}

		for _, item := range values {
			if item.value == "" {
				continue
			}

			if _, err := fmt.Fprintf(file, "%s=%s\n", item.key, item.value); err != nil {
				_ = file.Close()
				return fmt.Errorf("write account value: %w", err)
			}
		}

		for key, value := range account.Values {
			if key == "" || value == "" {
				continue
			}

			switch key {
			case "label", "client_id", "secret_id", "root_folder_id", "remote_type":
				continue
			}

			if _, err := fmt.Fprintf(file, "%s=%s\n", key, value); err != nil {
				_ = file.Close()
				return fmt.Errorf("write account value %q: %w", key, err)
			}
		}

		if _, err := file.WriteString("\n"); err != nil {
			_ = file.Close()
			return fmt.Errorf("write account separator: %w", err)
		}
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("close account file: %w", err)
	}

	return nil
}

func InitializeAccounts() error {
	return SaveAccounts([]Account{})
}

func AddAccount(account Account) error {
	if strings.TrimSpace(account.Label) == "" {
		return errors.New("account label cannot be empty")
	}

	accounts, err := LoadAccounts()
	if err != nil {
		return err
	}

	accounts = append(accounts, account)

	return SaveAccounts(accounts)
}

func UpdateAccount(index int, account Account) error {
	if strings.TrimSpace(account.Label) == "" {
		return errors.New("account label cannot be empty")
	}

	accounts, err := LoadAccounts()
	if err != nil {
		return err
	}

	if index < 0 || index >= len(accounts) {
		return fmt.Errorf("account index out of range: %d", index)
	}

	accounts[index] = account

	return SaveAccounts(accounts)
}

func DeleteAccount(index int) error {
	accounts, err := LoadAccounts()
	if err != nil {
		return err
	}

	if index < 0 || index >= len(accounts) {
		return fmt.Errorf("account index out of range: %d", index)
	}

	accounts = append(accounts[:index], accounts[index+1:]...)

	return SaveAccounts(accounts)
}

func GetAccount(index int) (*Account, error) {
	accounts, err := LoadAccounts()
	if err != nil {
		return nil, err
	}

	if index < 0 || index >= len(accounts) {
		return nil, fmt.Errorf("account index out of range: %d", index)
	}

	account := accounts[index]

	return &account, nil
}

func FindAccount(label string) (*Account, error) {
	accounts, err := LoadAccounts()
	if err != nil {
		return nil, err
	}

	for _, account := range accounts {
		if account.Label == label {
			return &account, nil
		}
	}

	return nil, fmt.Errorf("account not found: %s", label)
}

func ResetAccounts() error {
	path, err := AccountFilePath()
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return fmt.Errorf("remove account file: %w", err)
	}

	return nil
}

func AccountsConfigured() (bool, error) {
	accounts, err := LoadAccounts()
	if err != nil {
		return false, err
	}

	return len(accounts) > 0, nil
}