package exchange

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"acrux/internal/config"
	fileops "acrux/internal/file"

	tea "github.com/charmbracelet/bubbletea"
)

type fileOperationMode int

const (
	fileModeLocal fileOperationMode = iota
	fileModeSelect
	fileModeDestination
)

type fileEntry struct {
	name  string
	path  string
	isDir bool
}

type fileModel struct {
	mode       fileOperationMode
	cursor     int
	items      []fileEntry
	path       string
	selected   []string
	destination string
	message    string
	quitting   bool
	operation  string
}

func FileMenu(operation string) (string, error) {
	if strings.TrimSpace(operation) == "" {
		return "", errors.New("file operation cannot be empty")
	}

	preference, err := config.LoadPreference()
	if err != nil {
		return "", fmt.Errorf("load preferences: %w", err)
	}

	startPath := preference.StartingLocalDirectory
	if startPath == "" {
		startPath, err = os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("get home directory: %w", err)
		}
	}

	model := fileModel{
		mode:      fileModeLocal,
		path:      startPath,
		operation: operation,
	}

	model, err = model.loadDirectory()
	if err != nil {
		return "", err
	}

	result, err := tea.NewProgram(model).Run()
	if err != nil {
		return "", fmt.Errorf("run file interface: %w", err)
	}

	m, ok := result.(fileModel)
	if !ok {
		return "", errors.New("invalid file interface result")
	}

	return m.result(), nil
}

func (m fileModel) Init() tea.Cmd {
	return nil
}

func (m fileModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			m.message = "Cancelled"
			return m, tea.Quit

		case "esc":
			if m.mode == fileModeLocal {
				m.quitting = true
				m.message = "Cancelled"
				return m, tea.Quit
			}

			m.mode = fileModeLocal
			m.selected = nil
			m.destination = ""
			m.message = ""

			return m, nil

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}

		case "space":
			if m.mode == fileModeLocal {
				m.toggleSelection()
			}

		case "enter":
			return m.selectCurrent()
		}
	}

	return m, nil
}

func (m fileModel) View() string {
	if m.quitting {
		return ""
	}

	var builder strings.Builder

	builder.WriteString("File Operation: ")
	builder.WriteString(m.operation)
	builder.WriteString("\n\n")

	builder.WriteString("Path: ")
	builder.WriteString(m.path)
	builder.WriteString("\n\n")

	if m.mode == fileModeDestination {
		builder.WriteString("Destination:\n\n")
	} else {
		builder.WriteString("Select files/directories:\n\n")
	}

	for i, item := range m.items {
		cursor := " "

		if i == m.cursor {
			cursor = ">"
		}

		selected := " "
		if contains(m.selected, item.path) {
			selected = "*"
		}

		builder.WriteString(fmt.Sprintf(
			"%s [%s] %s",
			cursor,
			selected,
			item.name,
		))

		if item.isDir {
			builder.WriteString("/")
		}

		builder.WriteString("\n")
	}

	if len(m.selected) > 0 {
		builder.WriteString("\nSelected: ")
		builder.WriteString(fmt.Sprintf("%d item(s)", len(m.selected)))
	}

	if m.message != "" {
		builder.WriteString("\n\n")
		builder.WriteString(m.message)
	}

	builder.WriteString("\n\n↑/↓ navigate  Space select  Enter confirm  Esc back  q quit")

	return builder.String()
}

func (m fileModel) selectCurrent() (tea.Model, tea.Cmd) {
	if len(m.items) == 0 {
		return m, nil
	}

	item := m.items[m.cursor]

	if item.name == ".." {
		parent := filepath.Dir(m.path)

		if parent != m.path {
			m.path = parent
			m.cursor = 0

			updated, err := m.loadDirectory()
			if err != nil {
				m.message = err.Error()
				return m, nil
			}

			return updated, nil
		}

		return m, nil
	}

	if m.mode == fileModeDestination {
		if item.isDir {
			m.destination = item.path
			m.message = "Destination selected: " + item.path
			m.quitting = true

			return m, tea.Quit
		}

		m.message = "Select a directory as the destination."
		return m, nil
	}

	if item.isDir {
		m.path = item.path
		m.cursor = 0

		updated, err := m.loadDirectory()
		if err != nil {
			m.message = err.Error()
			return m, nil
		}

		return updated, nil
	}

	m.toggleSelection()

	return m, nil
}

func (m *fileModel) toggleSelection() {
	if m.cursor < 0 || m.cursor >= len(m.items) {
		return
	}

	item := m.items[m.cursor]

	if item.name == ".." {
		return
	}

	for i, selected := range m.selected {
		if selected == item.path {
			m.selected = append(m.selected[:i], m.selected[i+1:]...)
			return
		}
	}

	m.selected = append(m.selected, item.path)
}

func (m fileModel) loadDirectory() (fileModel, error) {
	entries, err := os.ReadDir(m.path)
	if err != nil {
		return m, fmt.Errorf("read directory %q: %w", m.path, err)
	}

	items := make([]fileEntry, 0, len(entries)+1)

	parent := filepath.Dir(m.path)
	if parent != m.path {
		items = append(items, fileEntry{
			name:  "..",
			path:  parent,
			isDir: true,
		})
	}

	for _, entry := range entries {
		entryPath := filepath.Join(m.path, entry.Name())

		info, err := entry.Info()
		if err != nil {
			continue
		}

		items = append(items, fileEntry{
			name:  entry.Name(),
			path:  entryPath,
			isDir: info.IsDir(),
		})
	}

	m.items = items

	if m.cursor >= len(m.items) {
		m.cursor = 0
	}

	return m, nil
}

func (m fileModel) result() string {
	switch {
	case m.message == "Cancelled":
		return ""

	case m.destination != "":
		return m.destination

	case len(m.selected) > 0:
		return strings.Join(m.selected, "\n")

	default:
		return ""
	}
}

func CopyLocal(source, destination string) error {
	if strings.TrimSpace(source) == "" {
		return errors.New("source cannot be empty")
	}

	if strings.TrimSpace(destination) == "" {
		return errors.New("destination cannot be empty")
	}

	if err := fileops.Copy(source, destination); err != nil {
		return fmt.Errorf("copy local file: %w", err)
	}

	return nil
}

func DeleteLocal(path string) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("path cannot be empty")
	}

	if err := fileops.Delete(path); err != nil {
		return fmt.Errorf("delete local file: %w", err)
	}

	return nil
}

func contains(values []string, value string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}

	return false
}