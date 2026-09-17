package exchange

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"acrux/internal/chunk"

	tea "github.com/charmbracelet/bubbletea"
)

type chunkMenuModel struct {
	title   string
	items   []string
	cursor  int
	selected string
	quitting bool
}

func newChunkMenuModel(title string, items []string) chunkMenuModel {
	return chunkMenuModel{
		title: title,
		items: items,
	}
}

func (m chunkMenuModel) Init() tea.Cmd {
	return nil
}

func (m chunkMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m chunkMenuModel) View() string {
	if m.quitting {
		return ""
	}

	var builder strings.Builder

	builder.WriteString(m.title)
	builder.WriteString("\n\n")

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

// ChunkMenu provides an interactive menu for splitting or joining files.
func ChunkMenu() (string, error) {
	model := newChunkMenuModel(
		"Chunk",
		[]string{
			"Split",
			"Join",
			"Back",
		},
	)

	result, err := tea.NewProgram(model).Run()
	if err != nil {
		return "", err
	}

	finalModel, ok := result.(chunkMenuModel)
	if !ok {
		return "", fmt.Errorf("invalid chunk menu result")
	}

	if finalModel.selected == "" || finalModel.selected == "Back" {
		return "", nil
	}

	return finalModel.selected, nil
}

// SplitFileIntoChunks splits source into numbered chunk files.
func SplitFileIntoChunks(source, destinationDir string, maxSize int64) ([]string, error) {
	if source == "" {
		return nil, fmt.Errorf("source file is required")
	}

	if destinationDir == "" {
		return nil, fmt.Errorf("destination directory is required")
	}

	if maxSize <= 0 {
		return nil, fmt.Errorf("maximum chunk size must be greater than zero")
	}

	if err := os.MkdirAll(destinationDir, 0755); err != nil {
		return nil, fmt.Errorf("create chunk directory: %w", err)
	}

	return chunk.Split(source, destinationDir, maxSize)
}

// JoinChunkFiles joins chunk files into destination.
func JoinChunkFiles(chunks []string, destination string) error {
	if len(chunks) == 0 {
		return fmt.Errorf("at least one chunk is required")
	}

	if destination == "" {
		return fmt.Errorf("destination file is required")
	}

	return chunk.Join(chunks, destination)
}

// FindChunks returns chunk files belonging to a source file.
//
// Chunk files are expected to use the format:
//
//	source.ext.part1
//	source.ext.part2
//	source.ext.part3
func FindChunks(directory, baseName string) ([]string, error) {
	if directory == "" {
		return nil, fmt.Errorf("directory is required")
	}

	if baseName == "" {
		return nil, fmt.Errorf("base file name is required")
	}

	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("read chunk directory: %w", err)
	}

	prefix := baseName + ".part"
	type numberedChunk struct {
		number int
		path   string
	}

	var found []numberedChunk

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasPrefix(name, prefix) {
			continue
		}

		numberText := strings.TrimPrefix(name, prefix)
		number, err := strconv.Atoi(numberText)
		if err != nil || number <= 0 {
			continue
		}

		found = append(found, numberedChunk{
			number: number,
			path:   filepath.Join(directory, name),
		})
	}

	sort.Slice(found, func(i, j int) bool {
		return found[i].number < found[j].number
	})

	result := make([]string, 0, len(found))
	for _, item := range found {
		result = append(result, item.path)
	}

	return result, nil
}

// SplitInteractive performs an interactive split operation.
func SplitInteractive(source, destinationDir string, maxSize int64) ([]string, error) {
	return SplitFileIntoChunks(source, destinationDir, maxSize)
}

// JoinInteractive performs a join operation using explicitly supplied chunks.
func JoinInteractive(chunks []string, destination string) error {
	return JoinChunkFiles(chunks, destination)
}