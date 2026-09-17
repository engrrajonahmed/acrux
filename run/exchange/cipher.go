package exchange

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"acrux/internal/cipher"

	tea "github.com/charmbracelet/bubbletea"
)

type cipherOperation int

const (
	cipherEncrypt cipherOperation = iota
	cipherDecrypt
)

type cipherModel struct {
	operation cipherOperation
	cursor    int
	items     []string
	choice    string
	quitting  bool
}

func CipherMenu() (string, error) {
	model := cipherModel{
		items: []string{
			"Encrypt",
			"Decrypt",
			"Back",
		},
	}

	result, err := tea.NewProgram(model).Run()
	if err != nil {
		return "", fmt.Errorf("run cipher menu: %w", err)
	}

	m, ok := result.(cipherModel)
	if !ok {
		return "", errors.New("invalid cipher menu result")
	}

	return m.choice, nil
}

func (m cipherModel) Init() tea.Cmd {
	return nil
}

func (m cipherModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.choice = "Back"
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
			if len(m.items) == 0 {
				m.choice = "Back"
				m.quitting = true
				return m, tea.Quit
			}

			m.choice = m.items[m.cursor]
			m.quitting = true

			return m, tea.Quit
		}
	}

	return m, nil
}

func (m cipherModel) View() string {
	if m.quitting {
		return ""
	}

	var builder strings.Builder

	builder.WriteString("Cipher\n\n")

	for i, item := range m.items {
		cursor := " "

		if i == m.cursor {
			cursor = ">"
		}

		builder.WriteString(fmt.Sprintf("%s %s\n", cursor, item))
	}

	builder.WriteString("\n↑/↓ navigate  Enter select  Esc back  q quit")

	return builder.String()
}

func EncryptFile(source, destination, key string) error {
	if strings.TrimSpace(source) == "" {
		return errors.New("source cannot be empty")
	}

	if strings.TrimSpace(destination) == "" {
		return errors.New("destination cannot be empty")
	}

	if key == "" {
		return errors.New("encryption key cannot be empty")
	}

	if err := cipher.Encrypt(source, destination, key); err != nil {
		return fmt.Errorf("encrypt file: %w", err)
	}

	return nil
}

func DecryptFile(source, destination, key string) error {
	if strings.TrimSpace(source) == "" {
		return errors.New("source cannot be empty")
	}

	if strings.TrimSpace(destination) == "" {
		return errors.New("destination cannot be empty")
	}

	if key == "" {
		return errors.New("encryption key cannot be empty")
	}

	if err := cipher.Decrypt(source, destination, key); err != nil {
		return fmt.Errorf("decrypt file: %w", err)
	}

	return nil
}

func EncryptReader(reader io.Reader, writer io.Writer, key string) error {
	if reader == nil {
		return errors.New("reader cannot be nil")
	}

	if writer == nil {
		return errors.New("writer cannot be nil")
	}

	if key == "" {
		return errors.New("encryption key cannot be empty")
	}

	if err := cipher.EncryptReader(reader, writer, key); err != nil {
		return fmt.Errorf("encrypt stream: %w", err)
	}

	return nil
}

func DecryptReader(reader io.Reader, writer io.Writer, key string) error {
	if reader == nil {
		return errors.New("reader cannot be nil")
	}

	if writer == nil {
		return errors.New("writer cannot be nil")
	}

	if key == "" {
		return errors.New("encryption key cannot be empty")
	}

	if err := cipher.DecryptReader(reader, writer, key); err != nil {
		return fmt.Errorf("decrypt stream: %w", err)
	}

	return nil
}