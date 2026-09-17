package exchange

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type cipherModel struct {
	items    []string
	cursor   int
	selected string
}

func NewCipher() *cipherModel {
	return &cipherModel{
		items: []string{
			"Encrypt",
			"Decrypt",
			"Back",
		},
	}
}

func (m cipherModel) Init() tea.Cmd {
	return nil
}

func (m cipherModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m cipherModel) View() string {
	var b strings.Builder

	b.WriteString("Cipher\n\n")

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

func CipherMenu() (string, error) {
	model := NewCipher()

	result, err := tea.NewProgram(
		model,
		tea.WithAltScreen(),
	).Run()
	if err != nil {
		return "", err
	}

	finalModel, ok := result.(cipherModel)
	if !ok {
		return "", nil
	}

	return finalModel.selected, nil
}