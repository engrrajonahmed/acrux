package exchange

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type hashModel struct {
	items    []string
	cursor   int
	selected string
}

func NewHash() *hashModel {
	return &hashModel{
		items: []string{
			"Calculate",
			"Compare",
			"Back",
		},
	}
}

func (m hashModel) Init() tea.Cmd {
	return nil
}

func (m hashModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m hashModel) View() string {
	var b strings.Builder

	b.WriteString("Hash\n\n")

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

func HashMenu() (string, error) {
	model := NewHash()

	result, err := tea.NewProgram(
		model,
		tea.WithAltScreen(),
	).Run()
	if err != nil {
		return "", err
	}

	finalModel, ok := result.(hashModel)
	if !ok {
		return "", nil
	}

	return finalModel.selected, nil
}