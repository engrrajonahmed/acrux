package exchange

import (
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"acrux/internal/config"
)

type manageItem struct {
	title string
}

type manageModel struct {
	items    []manageItem
	cursor   int
	choice   string
	quitting bool
}

func ManageMenu() (string, error) {
	items := []manageItem{
		{title: "Accounts"},
		{title: "Settings"},
		{title: "Reset"},
		{title: "Uninstall"},
		{title: "Back"},
	}

	model := manageModel{
		items: items,
	}

	result, err := tea.NewProgram(model).Run()
	if err != nil {
		return "", fmt.Errorf("run manage menu: %w", err)
	}

	m, ok := result.(manageModel)
	if !ok {
		return "", errors.New("invalid manage menu result")
	}

	return m.choice, nil
}

func (m manageModel) Init() tea.Cmd {
	return nil
}

func (m manageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
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

			m.choice = m.items[m.cursor].title
			m.quitting = true

			return m, tea.Quit
		}
	}

	return m, nil
}

func (m manageModel) View() string {
	if m.quitting {
		return ""
	}

	view := "Manage\n\n"

	for i, item := range m.items {
		cursor := " "

		if m.cursor == i {
			cursor = ">"
		}

		view += fmt.Sprintf("%s %s\n", cursor, item.title)
	}

	view += "\nUse ↑/↓ to navigate and Enter to select."

	return view
}

func ResetConfiguration() error {
	accounts, err := config.LoadAccounts()
	if err != nil {
		return fmt.Errorf("load accounts: %w", err)
	}

	preference, err := config.LoadPreference()
	if err != nil {
		return fmt.Errorf("load preferences: %w", err)
	}

	if accounts == nil {
		accounts = []config.Account{}
	}

	if preference == nil {
		preference, err = config.DefaultPreference()
		if err != nil {
			return fmt.Errorf("create default preferences: %w", err)
		}
	}

	if err := config.ResetAccounts(); err != nil {
		return fmt.Errorf("reset account configuration: %w", err)
	}

	if err := config.ResetPreference(); err != nil {
		return fmt.Errorf("reset preference configuration: %w", err)
	}

	return nil
}