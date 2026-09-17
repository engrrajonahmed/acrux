package exchange

import (
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type MenuItem struct {
	Title string
}

type menuModel struct {
	items  []MenuItem
	cursor int
	choice string
	quitting bool
}

func MainMenu() (string, error) {
	items := []MenuItem{
		{Title: "Browse"},
		{Title: "Copy"},
		{Title: "Delete"},
		{Title: "Manage"},
		{Title: "Exit"},
	}

	model := menuModel{
		items: items,
	}

	result, err := tea.NewProgram(model).Run()
	if err != nil {
		return "", fmt.Errorf("run main menu: %w", err)
	}

	m, ok := result.(menuModel)
	if !ok {
		return "", errors.New("invalid main menu result")
	}

	return m.choice, nil
}

func (m menuModel) Init() tea.Cmd {
	return nil
}

func (m menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.choice = "Exit"
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
				m.choice = "Exit"
				m.quitting = true
				return m, tea.Quit
			}

			m.choice = m.items[m.cursor].Title

			if m.choice == "Exit" {
				m.quitting = true
				return m, tea.Quit
			}

			return m, tea.Quit
		}
	}

	return m, nil
}

func (m menuModel) View() string {
	if m.quitting {
		return ""
	}

	var view string

	view += "acrux\n\n"

	for i, item := range m.items {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		view += fmt.Sprintf("%s %s\n", cursor, item.Title)
	}

	view += "\nUse ↑/↓ to navigate and Enter to select."

	return view
}