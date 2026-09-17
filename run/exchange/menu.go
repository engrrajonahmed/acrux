package exchange

import (
	tea "github.com/charmbracelet/bubbletea"
)

type menuModel struct {
	cursor int
	items  []string
}

func NewMenu() *menuModel {
	return &menuModel{
		items: []string{
			"Browse",
			"Copy",
			"Delete",
			"Manage",
			"Exit",
		},
	}
}

func (m menuModel) Init() tea.Cmd {
	return nil
}

func (m menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			switch m.cursor {
			case 0:
				return m, func() tea.Msg {
					return BrowseMsg{}
				}
			case 1:
				return m, func() tea.Msg {
					return CopyMsg{}
				}
			case 2:
				return m, func() tea.Msg {
					return DeleteMsg{}
				}
			case 3:
				return m, func() tea.Msg {
					return ManageMsg{}
				}
			case 4:
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m menuModel) View() string {
	s := "acrux\n\n"

	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}

		s += cursor + item + "\n"
	}

	s += "\nUse ↑/↓ to navigate and Enter to select."

	return s
}

type BrowseMsg struct{}
type CopyMsg struct{}
type DeleteMsg struct{}
type ManageMsg struct{}

func MainMenu() error {
	model := NewMenu()

	_, err := tea.NewProgram(
		model,
		tea.WithAltScreen(),
	).Run()

	return err
}