package exchange

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"acrux/internal/hash"

	tea "github.com/charmbracelet/bubbletea"
)

type hashMenuModel struct {
	items    []string
	cursor   int
	selected string
	quitting bool
}

func newHashMenuModel() hashMenuModel {
	return hashMenuModel{
		items: []string{
			"Calculate Hash",
			"Compare Hash",
			"Back",
		},
	}
}

func (m hashMenuModel) Init() tea.Cmd {
	return nil
}

func (m hashMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}

		case "enter", " ":
			if len(m.items) > 0 {
				m.selected = m.items[m.cursor]
				m.quitting = true
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m hashMenuModel) View() string {
	if m.quitting {
		return ""
	}

	var builder strings.Builder

	builder.WriteString("Hash\n\n")

	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}

		builder.WriteString(cursor)
		builder.WriteString(item)
		builder.WriteString("\n")
	}

	builder.WriteString("\nUse ↑/↓ to navigate, Enter to select, Esc to cancel.\n")

	return builder.String()
}

// HashMenu provides an interactive menu for hash operations.
func HashMenu() (string, error) {
	model := newHashMenuModel()

	result, err := tea.NewProgram(model).Run()
	if err != nil {
		return "", err
	}

	finalModel, ok := result.(hashMenuModel)
	if !ok {
		return "", fmt.Errorf("invalid hash menu result")
	}

	if finalModel.selected == "" || finalModel.selected == "Back" {
		return "", nil
	}

	return finalModel.selected, nil
}

// CalculateFileHash calculates the SHA-256 hash of a file.
func CalculateFileHash(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("file path is required")
	}

	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("stat file: %w", err)
	}

	if info.IsDir() {
		return "", fmt.Errorf("cannot calculate hash of directory: %s", path)
	}

	return hash.Calculate(path)
}

// CalculateFilesHash calculates SHA-256 hashes for multiple files.
func CalculateFilesHash(paths []string) (map[string]string, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("at least one file is required")
	}

	results := make(map[string]string, len(paths))

	for _, path := range paths {
		digest, err := CalculateFileHash(path)
		if err != nil {
			return nil, err
		}

		results[path] = digest
	}

	return results, nil
}

// CompareFileHashes compares the contents of two files using SHA-256.
func CompareFileHashes(first, second string) (bool, error) {
	if first == "" {
		return false, fmt.Errorf("first file path is required")
	}

	if second == "" {
		return false, fmt.Errorf("second file path is required")
	}

	firstInfo, err := os.Stat(first)
	if err != nil {
		return false, fmt.Errorf("stat first file: %w", err)
	}

	secondInfo, err := os.Stat(second)
	if err != nil {
		return false, fmt.Errorf("stat second file: %w", err)
	}

	if firstInfo.IsDir() {
		return false, fmt.Errorf("first path is a directory: %s", first)
	}

	if secondInfo.IsDir() {
		return false, fmt.Errorf("second path is a directory: %s", second)
	}

	return hash.CompareFiles(first, second)
}

// CompareHashValues compares two hexadecimal SHA-256 hash values.
func CompareHashValues(first, second string) bool {
	return strings.EqualFold(
		strings.TrimSpace(first),
		strings.TrimSpace(second),
	)
}

// HashTemporaryFile calculates a hash for a temporary file and removes it.
func HashTemporaryFile(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("file path is required")
	}

	digest, err := CalculateFileHash(path)
	if removeErr := os.Remove(path); err == nil && removeErr != nil {
		return "", fmt.Errorf("remove temporary file: %w", removeErr)
	}

	if err != nil {
		return "", err
	}

	return digest, nil
}

// HashFileName returns a hash file path next to the supplied file.
func HashFileName(path string) string {
	if path == "" {
		return ""
	}

	return filepath.Join(
		filepath.Dir(path),
		filepath.Base(path)+".sha256",
	)
}

// WriteFileHash calculates a file hash and writes it to a .sha256 file.
func WriteFileHash(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("file path is required")
	}

	digest, err := CalculateFileHash(path)
	if err != nil {
		return "", err
	}

	hashPath := HashFileName(path)

	content := fmt.Sprintf("%s  %s\n", digest, filepath.Base(path))
	if err := os.WriteFile(hashPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("write hash file: %w", err)
	}

	return hashPath, nil
}