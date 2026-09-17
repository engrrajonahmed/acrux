package exchange

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type codecModel struct {
	items    []string
	cursor   int
	selected string
}

func NewCodec() *codecModel {
	return &codecModel{
		items: []string{
			"Compress",
			"Decompress",
			"Back",
		},
	}
}

func (m codecModel) Init() tea.Cmd {
	return nil
}

func (m codecModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
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
				return m, tea.Quit
			}

			m.selected = m.items[m.cursor]
			return m, tea.Quit

		case "esc", "q":
			m.selected = "Back"
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m codecModel) View() string {
	var b strings.Builder

	b.WriteString("Codec\n\n")

	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}

		b.WriteString(cursor)
		b.WriteString(item)
		b.WriteByte('\n')
	}

	b.WriteString("\n↑/↓ navigate, Enter select, q/esc back\n")

	return b.String()
}

func CodecMenu() (string, error) {
	model := NewCodec()

	result, err := tea.NewProgram(
		model,
		tea.WithAltScreen(),
	).Run()
	if err != nil {
		return "", err
	}

	finalModel, ok := result.(codecModel)
	if !ok {
		return "", nil
	}

	return finalModel.selected, nil
}