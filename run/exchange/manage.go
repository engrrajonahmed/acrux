package exchange

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type manageModel struct {
	items    []string
	cursor   int
	selected string
}

func NewManage() *manageModel {
	return &manageModel{
		items: []string{
			"Accounts",
			"Settings",
			"Reset",
			"Uninstall",
			"Back",
		},
	}
}

func (m manageModel) Init() tea.Cmd {
	return nil
}

func (m manageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m manageModel) View() string {
	var b strings.Builder

	b.WriteString("Manage\n\n")

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

func ManageMenu() (string, error) {
	model := NewManage()

	result, err := tea.NewProgram(
		model,
		tea.WithAltScreen(),
	).Run()
	if err != nil {
		return "", err
	}

	finalModel, ok := result.(manageModel)
	if !ok {
		return "", nil
	}

	return finalModel.selected, nil
}