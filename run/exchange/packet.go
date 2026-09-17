package exchange

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"acrux/internal/chunk"
	"acrux/internal/cipher"
	"acrux/internal/codec"
	"acrux/internal/config"
	"acrux/internal/packet"

	tea "github.com/charmbracelet/bubbletea"
)

type packetModel struct {
	items    []string
	cursor   int
	selected []string
	message  string
	quitting bool
}

func PacketMenu() ([]string, error) {
	model := packetModel{
		items: []string{
			"Create Archive",
			"Export Archive",
			"Back",
		},
	}

	result, err := tea.NewProgram(model).Run()
	if err != nil {
		return nil, fmt.Errorf("run packet menu: %w", err)
	}

	m, ok := result.(packetModel)
	if !ok {
		return nil, errors.New("invalid packet menu result")
	}

	if m.message == "Cancelled" || m.quitting && len(m.selected) == 0 {
		return nil, nil
	}

	return m.selected, nil
}

func (m packetModel) Init() tea.Cmd {
	return nil
}

func (m packetModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.message = "Cancelled"
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

		case "enter":
			if len(m.items) == 0 {
				m.quitting = true
				return m, tea.Quit
			}

			switch m.items[m.cursor] {
			case "Create Archive":
				m.selected = append(m.selected, "Create Archive")
				m.quitting = true
				return m, tea.Quit

			case "Export Archive":
				m.selected = append(m.selected, "Export Archive")
				m.quitting = true
				return m, tea.Quit

			case "Back":
				m.quitting = true
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m packetModel) View() string {
	if m.quitting {
		return ""
	}

	var builder strings.Builder

	builder.WriteString("Packet\n\n")

	for i, item := range m.items {
		cursor := " "

		if i == m.cursor {
			cursor = ">"
		}

		builder.WriteString(fmt.Sprintf("%s %s\n", cursor, item))
	}

	if m.message != "" {
		builder.WriteString("\n")
		builder.WriteString(m.message)
	}

	builder.WriteString("\n\n↑/↓ navigate  Enter select  Esc back  q quit")

	return builder.String()
}

func CreatePacket(sources []string, destination string) error {
	if len(sources) == 0 {
		return errors.New("at least one source is required")
	}

	if strings.TrimSpace(destination) == "" {
		return errors.New("archive destination cannot be empty")
	}

	if err := packet.Archive(sources, destination); err != nil {
		return fmt.Errorf("create archive: %w", err)
	}

	return nil
}

func ExportPacket(source, destination string) error {
	if strings.TrimSpace(source) == "" {
		return errors.New("archive source cannot be empty")
	}

	if strings.TrimSpace(destination) == "" {
		return errors.New("export destination cannot be empty")
	}

	if err := packet.Export(source, destination); err != nil {
		return fmt.Errorf("export archive: %w", err)
	}

	return nil
}

func PrepareSecurePacket(
	sources []string,
	workingDirectory string,
) (string, error) {
	if len(sources) == 0 {
		return "", errors.New("at least one source is required")
	}

	if strings.TrimSpace(workingDirectory) == "" {
		return "", errors.New("working directory cannot be empty")
	}

	if err := os.MkdirAll(workingDirectory, 0700); err != nil {
		return "", fmt.Errorf("create working directory: %w", err)
	}

	archivePath := filepath.Join(workingDirectory, "acrux.tar")
	compressedPath := filepath.Join(workingDirectory, "acrux.tar.gz")
	encryptedPath := filepath.Join(workingDirectory, "acrux.tar.gz.enc")

	if err := packet.Archive(sources, archivePath); err != nil {
		return "", fmt.Errorf("archive secure packet: %w", err)
	}

	if err := codec.Compress(archivePath, compressedPath); err != nil {
		return "", fmt.Errorf("compress secure packet: %w", err)
	}

	preference, err := config.LoadPreference()
	if err != nil {
		return "", fmt.Errorf("load preferences: %w", err)
	}

	if strings.TrimSpace(preference.EncryptionKey) == "" {
		return "", errors.New("encryption key is not configured")
	}

	if err := cipher.Encrypt(
		compressedPath,
		encryptedPath,
		preference.EncryptionKey,
	); err != nil {
		return "", fmt.Errorf("encrypt secure packet: %w", err)
	}

	return encryptedPath, nil
}

func SplitSecurePacket(
	encryptedPacket string,
	destination string,
) ([]string, error) {
	if strings.TrimSpace(encryptedPacket) == "" {
		return nil, errors.New("encrypted packet cannot be empty")
	}

	if strings.TrimSpace(destination) == "" {
		return nil, errors.New("chunk destination cannot be empty")
	}

	preference, err := config.LoadPreference()
	if err != nil {
		return nil, fmt.Errorf("load preferences: %w", err)
	}

	if preference.MaxArchiveSize <= 0 {
		return nil, errors.New("maximum archive size must be greater than zero")
	}

	info, err := os.Stat(encryptedPacket)
	if err != nil {
		return nil, fmt.Errorf("stat encrypted packet: %w", err)
	}

	if info.Size() <= preference.MaxArchiveSize {
		return []string{encryptedPacket}, nil
	}

	chunks, err := chunk.Split(
		encryptedPacket,
		destination,
		preference.MaxArchiveSize,
	)
	if err != nil {
		return nil, fmt.Errorf("split secure packet: %w", err)
	}

	return chunks, nil
}