package exchange

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"acrux/internal/browse"
	"acrux/internal/config"

	tea "github.com/charmbracelet/bubbletea"
)

type browseModel struct {
	items    []browse.LocalItem
	cloud    []browse.CloudItem
	isCloud  bool
	account  config.Account
	path     string
	cursor   int
	selected string
	quitting bool
	err      error
}

func newBrowseLocalModel(startPath string) browseModel {
	return browseModel{
		path: startPath,
	}
}

func newBrowseCloudModel(account config.Account, remotePath string) browseModel {
	return browseModel{
		isCloud: true,
		account: account,
		path:    remotePath,
	}
}

func (m browseModel) Init() tea.Cmd {
	return func() tea.Msg {
		if m.isCloud {
			items, err := browse.ListCloud(m.account, m.path)
			return cloudBrowseLoadedMsg{
				items: items,
				err:   err,
			}
		}

		items, err := browse.ListLocal(m.path)
		return localBrowseLoadedMsg{
			items: items,
			err:   err,
		}
	}
}

type localBrowseLoadedMsg struct {
	items []browse.LocalItem
	err   error
}

type cloudBrowseLoadedMsg struct {
	items []browse.CloudItem
	err   error
}

func (m browseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case localBrowseLoadedMsg:
		m.items = msg.items
		m.err = msg.err
		return m, nil

	case cloudBrowseLoadedMsg:
		m.cloud = msg.items
		m.err = msg.err
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "esc":
			if m.path == "" || (!m.isCloud && isHomePath(m.path)) {
				m.quitting = true
				return m, tea.Quit
			}

			m.path = browseParentPath(m.path, m.isCloud)
			m.cursor = 0
			m.err = nil
			return m, m.Init()

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			count := m.itemCount()
			if m.cursor < count-1 {
				m.cursor++
			}

		case "enter", "right", "l":
			if m.itemCount() == 0 {
				return m, nil
			}

			if m.isCloud {
				item := m.cloud[m.cursor]
				if item.IsDir {
					m.path = strings.Trim(item.Path, "/")
					m.cursor = 0
					m.err = nil
					return m, m.Init()
				}
			} else {
				item := m.items[m.cursor]
				if item.IsDir {
					m.path = item.Path
					m.cursor = 0
					m.err = nil
					return m, m.Init()
				}
			}

		case "left", "h":
			if m.path == "" || (!m.isCloud && isHomePath(m.path)) {
				return m, nil
			}

			m.path = browseParentPath(m.path, m.isCloud)
			m.cursor = 0
			m.err = nil
			return m, m.Init()
		}
	}

	return m, nil
}

func (m browseModel) View() string {
	if m.quitting {
		return ""
	}

	var builder strings.Builder

	builder.WriteString("Browse\n\n")

	if m.isCloud {
		builder.WriteString("Account: ")
		builder.WriteString(m.account.Label)
		builder.WriteString("\n")
	}

	builder.WriteString("Path: ")
	if m.path == "" {
		builder.WriteString("/")
	} else {
		builder.WriteString(m.path)
	}
	builder.WriteString("\n\n")

	if m.err != nil {
		builder.WriteString("Error: ")
		builder.WriteString(m.err.Error())
		builder.WriteString("\n\n")
	}

	if m.isCloud {
		for i, item := range m.cloud {
			cursor := "  "
			if i == m.cursor {
				cursor = "> "
			}

			name := item.Name
			if item.IsDir {
				name += "/"
			}

			builder.WriteString(cursor)
			builder.WriteString(name)
			builder.WriteString("\n")
		}
	} else {
		for i, item := range m.items {
			cursor := "  "
			if i == m.cursor {
				cursor = "> "
			}

			name := item.Name
			if item.IsDir {
				name += "/"
			}

			builder.WriteString(cursor)
			builder.WriteString(name)
			builder.WriteString("\n")
		}
	}

	builder.WriteString("\n↑/↓ navigate, Enter open directory, Esc/← parent, q quit\n")

	return builder.String()
}

func (m browseModel) itemCount() int {
	if m.isCloud {
		return len(m.cloud)
	}

	return len(m.items)
}

func BrowseMenu() (string, error) {
	model := newBrowseTypeModel()

	result, err := tea.NewProgram(model).Run()
	if err != nil {
		return "", err
	}

	finalModel, ok := result.(browseTypeModel)
	if !ok {
		return "", fmt.Errorf("invalid browse menu result")
	}

	if finalModel.selected == "" || finalModel.selected == "Back" {
		return "", nil
	}

	switch finalModel.selected {
	case "Local":
		return browseLocal(), nil
	case "Cloud":
		return browseCloud(), nil
	default:
		return "", nil
	}
}

type browseTypeModel struct {
	items    []string
	cursor   int
	selected string
	quitting bool
}

func newBrowseTypeModel() browseTypeModel {
	return browseTypeModel{
		items: []string{
			"Local",
			"Cloud",
			"Back",
		},
	}
}

func (m browseTypeModel) Init() tea.Cmd {
	return nil
}

func (m browseTypeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			m.selected = m.items[m.cursor]
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m browseTypeModel) View() string {
	if m.quitting {
		return ""
	}

	var builder strings.Builder

	builder.WriteString("Browse\n\n")

	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}

		builder.WriteString(cursor)
		builder.WriteString(item)
		builder.WriteString("\n")
	}

	builder.WriteString("\n↑/↓ navigate, Enter select, Esc cancel\n")

	return builder.String()
}

func browseLocal() string {
	preference, err := config.LoadPreference()
	if err != nil {
		return err.Error()
	}

	startPath := preference.StartingLocalDirectory
	if startPath == "" {
		defaultPreference, defaultErr := config.DefaultPreference()
		if defaultErr != nil {
			return defaultErr.Error()
		}

		startPath = defaultPreference.StartingLocalDirectory
	}

	model := newBrowseLocalModel(startPath)

	result, err := tea.NewProgram(model).Run()
	if err != nil {
		return err.Error()
	}

	finalModel, ok := result.(browseModel)
	if !ok {
		return "invalid local browse result"
	}

	if finalModel.err != nil {
		return finalModel.err.Error()
	}

	return finalModel.selected
}

func browseCloud() string {
	accounts, err := config.LoadAccounts()
	if err != nil {
		return err.Error()
	}

	if len(accounts) == 0 {
		return "no cloud accounts configured"
	}

	accountItems := make([]string, 0, len(accounts))
	for _, account := range accounts {
		accountItems = append(accountItems, account.Label)
	}

	model := newAccountSelectionModel(accountItems)

	result, err := tea.NewProgram(model).Run()
	if err != nil {
		return err.Error()
	}

	finalModel, ok := result.(accountSelectionModel)
	if !ok || finalModel.selected == "" {
		return ""
	}

	var account config.Account
	for _, candidate := range accounts {
		if candidate.Label == finalModel.selected {
			account = candidate
			break
		}
	}

	cloudModel := newBrowseCloudModel(account, "")

	result, err = tea.NewProgram(cloudModel).Run()
	if err != nil {
		return err.Error()
	}

	finalCloudModel, ok := result.(browseModel)
	if !ok {
		return "invalid cloud browse result"
	}

	if finalCloudModel.err != nil {
		return finalCloudModel.err.Error()
	}

	return finalCloudModel.selected
}

type accountSelectionModel struct {
	items    []string
	cursor   int
	selected string
	quitting bool
}

func newAccountSelectionModel(items []string) accountSelectionModel {
	return accountSelectionModel{
		items: items,
	}
}

func (m accountSelectionModel) Init() tea.Cmd {
	return nil
}

func (m accountSelectionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if len(m.items) == 0 {
				m.quitting = true
				return m, tea.Quit
			}

			m.selected = m.items[m.cursor]
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m accountSelectionModel) View() string {
	if m.quitting {
		return ""
	}

	var builder strings.Builder

	builder.WriteString("Select Cloud Account\n\n")

	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}

		builder.WriteString(cursor)
		builder.WriteString(item)
		builder.WriteString("\n")
	}

	builder.WriteString("\n↑/↓ navigate, Enter select, Esc cancel\n")

	return builder.String()
}

func browseParentPath(current string, isCloud bool) string {
	if current == "" {
		return ""
	}

	if isCloud {
		current = strings.Trim(current, "/")
		if current == "" {
			return ""
		}

		index := strings.LastIndex(current, "/")
		if index < 0 {
			return ""
		}

		return current[:index]
	}

	parent := filepath.Dir(current)
	if parent == "." {
		return ""
	}

	return parent
}

func isHomePath(path string) bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	return filepath.Clean(path) == filepath.Clean(home)
}