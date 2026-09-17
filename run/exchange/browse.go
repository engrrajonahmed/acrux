package exchange

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"acrux/internal/browse"
	"acrux/internal/config"

	tea "github.com/charmbracelet/bubbletea"
)

type browseMode int

const (
	browseModeLocation browseMode = iota
	browseModeLocal
	browseModeCloudAccount
	browseModeCloud
)

type browseModel struct {
	mode       browseMode
	cursor     int
	items      []browseEntry
	path       string
	remote     string
	remotePath string
	choice     string
	message    string
	quitting   bool
}

type browseEntry struct {
	name     string
	path     string
	isDir    bool
	size     int64
	created  time.Time
	modified time.Time
}

func BrowseMenu() (string, error) {
	model := newBrowseModel()

	result, err := tea.NewProgram(model).Run()
	if err != nil {
		return "", fmt.Errorf("run browse interface: %w", err)
	}

	m, ok := result.(browseModel)
	if !ok {
		return "", errors.New("invalid browse result")
	}

	return m.choice, nil
}

func newBrowseModel() browseModel {
	return browseModel{
		mode: browseModeLocation,
		items: []browseEntry{
			{name: "Local"},
			{name: "Cloud"},
			{name: "Back"},
		},
	}
}

func (m browseModel) Init() tea.Cmd {
	return nil
}

func (m browseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.choice = "Back"
			m.quitting = true
			return m, tea.Quit

		case "esc", "backspace":
			return m.goBack()

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}

		case "enter", " ":
			return m.selectCurrent()
		}
	}

	return m, nil
}

func (m browseModel) View() string {
	if m.quitting {
		return ""
	}

	var builder strings.Builder

	switch m.mode {
	case browseModeLocation:
		builder.WriteString("Browse\n\n")

	case browseModeLocal:
		builder.WriteString("Browse / Local\n\n")
		builder.WriteString("Path: ")
		builder.WriteString(m.path)
		builder.WriteString("\n\n")

	case browseModeCloudAccount:
		builder.WriteString("Browse / Cloud\n\n")
		builder.WriteString("Select account:\n\n")

	case browseModeCloud:
		builder.WriteString("Browse / Cloud\n\n")
		builder.WriteString("Remote: ")
		builder.WriteString(m.remote)
		builder.WriteString("\nPath: ")
		builder.WriteString(m.remotePath)
		builder.WriteString("\n\n")
	}

	if m.message != "" {
		builder.WriteString(m.message)
		builder.WriteString("\n\n")
	}

	for i, item := range m.items {
		cursor := " "

		if i == m.cursor {
			cursor = ">"
		}

		builder.WriteString(fmt.Sprintf("%s %s", cursor, item.name))

		if m.mode == browseModeLocal ||
			m.mode == browseModeCloud {
			builder.WriteString(formatEntryMetadata(item))
		}

		builder.WriteString("\n")
	}

	builder.WriteString("\n↑/↓ navigate  Enter open  Esc back  q quit")

	return builder.String()
}

func (m browseModel) selectCurrent() (tea.Model, tea.Cmd) {
	if len(m.items) == 0 {
		return m, nil
	}

	entry := m.items[m.cursor]

	switch m.mode {
	case browseModeLocation:
		switch entry.name {
		case "Local":
			return m.openLocal()

		case "Cloud":
			return m.openCloudAccounts()

		case "Back":
			m.choice = "Back"
			m.quitting = true
			return m, tea.Quit
		}

	case browseModeLocal:
		if entry.name == ".." {
			parent := filepath.Dir(m.path)

			if parent != m.path {
				m.path = parent
				return m.reloadLocal()
			}

			return m, nil
		}

		if entry.isDir {
			m.path = entry.path
			return m.reloadLocal()
		}

	case browseModeCloudAccount:
		if entry.name == "Back" {
			return m.goBack()
		}

		m.remote = entry.name
		m.remotePath = ""

		return m.reloadCloud()

	case browseModeCloud:
		if entry.name == ".." {
			m.remotePath = browse.CloudParent(m.remotePath)
			return m.reloadCloud()
		}

		if entry.isDir {
			m.remotePath = entry.path
			return m.reloadCloud()
		}
	}

	return m, nil
}

func (m browseModel) openLocal() (tea.Model, tea.Cmd) {
	preference, err := config.LoadPreference()
	if err != nil {
		m.message = fmt.Sprintf("Unable to load preferences: %v", err)
		return m, nil
	}

	localPath := preference.StartingLocalDirectory

	if localPath == "" {
		localPath, err = userHomeDirectory()
		if err != nil {
			m.message = err.Error()
			return m, nil
		}
	}

	m.mode = browseModeLocal
	m.path = localPath
	m.cursor = 0
	m.message = ""

	return m.reloadLocal()
}

func (m browseModel) reloadLocal() (tea.Model, tea.Cmd) {
	items, err := browse.ListLocal(m.path)
	if err != nil {
		m.message = fmt.Sprintf("Unable to browse %q: %v", m.path, err)
		return m, nil
	}

	entries := make([]browseEntry, 0, len(items)+1)

	parent := filepath.Dir(m.path)

	if parent != m.path {
		entries = append(entries, browseEntry{
			name: "..",
			path: parent,
			isDir: true,
		})
	}

	for _, item := range items {
		entries = append(entries, browseEntry{
			name:     item.Name,
			path:     item.Path,
			isDir:    item.IsDir,
			size:     item.Size,
			created:  item.Created,
			modified: item.Modified,
		})
	}

	m.items = entries
	m.cursor = 0
	m.message = ""

	return m, nil
}

func (m browseModel) openCloudAccounts() (tea.Model, tea.Cmd) {
	accounts, err := config.LoadAccounts()
	if err != nil {
		m.message = fmt.Sprintf("Unable to load accounts: %v", err)
		return m, nil
	}

	if len(accounts) == 0 {
		m.mode = browseModeCloudAccount
		m.items = []browseEntry{
			{name: "Back"},
		}
		m.cursor = 0
		m.message = "No cloud accounts are configured."

		return m, nil
	}

	entries := make([]browseEntry, 0, len(accounts)+1)

	for _, account := range accounts {
		if strings.TrimSpace(account.Label) == "" {
			continue
		}

		entries = append(entries, browseEntry{
			name: account.Label,
		})
	}

	entries = append(entries, browseEntry{name: "Back"})

	m.mode = browseModeCloudAccount
	m.items = entries
	m.cursor = 0
	m.message = ""

	return m, nil
}

func (m browseModel) reloadCloud() (tea.Model, tea.Cmd) {
	items, err := browse.ListCloud(m.remote, m.remotePath)
	if err != nil {
		m.message = fmt.Sprintf(
			"Unable to browse %s:%s: %v",
			m.remote,
			m.remotePath,
			err,
		)
		return m, nil
	}

	entries := make([]browseEntry, 0, len(items)+1)

	if m.remotePath != "" {
		entries = append(entries, browseEntry{
			name: "..",
			path: browse.CloudParent(m.remotePath),
			isDir: true,
		})
	}

	for _, item := range items {
		entries = append(entries, browseEntry{
			name:     item.Name,
			path:     item.Path,
			isDir:    item.IsDir,
			size:     item.Size,
			created:  item.Created,
			modified: item.Modified,
		})
	}

	m.mode = browseModeCloud
	m.items = entries
	m.cursor = 0
	m.message = ""

	return m, nil
}

func (m browseModel) goBack() (tea.Model, tea.Cmd) {
	switch m.mode {
	case browseModeLocation:
		m.choice = "Back"
		m.quitting = true
		return m, tea.Quit

	case browseModeLocal, browseModeCloudAccount:
		m.mode = browseModeLocation
		m.items = []browseEntry{
			{name: "Local"},
			{name: "Cloud"},
			{name: "Back"},
		}
		m.cursor = 0
		m.message = ""
		return m, nil

	case browseModeCloud:
		m.mode = browseModeCloudAccount
		m.remote = ""
		m.remotePath = ""
		m.cursor = 0
		m.message = ""

		return m.openCloudAccounts()
	}

	return m, nil
}

func formatEntryMetadata(entry browseEntry) string {
	var parts []string

	if entry.isDir {
		parts = append(parts, "dir")
	} else {
		parts = append(parts, formatSize(entry.size))
	}

	if !entry.created.IsZero() {
		parts = append(parts, "created "+entry.created.Format("2006-01-02 15:04:05"))
	}

	if !entry.modified.IsZero() {
		parts = append(parts, "modified "+entry.modified.Format("2006-01-02 15:04:05"))
	}

	return "  [" + strings.Join(parts, ", ") + "]"
}

func formatSize(size int64) string {
	if size < 0 {
		return "unknown"
	}

	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)

	switch {
	case size >= gb:
		return fmt.Sprintf("%.2f GB", float64(size)/float64(gb))

	case size >= mb:
		return fmt.Sprintf("%.2f MB", float64(size)/float64(mb))

	case size >= kb:
		return fmt.Sprintf("%.2f KB", float64(size)/float64(kb))

	default:
		return fmt.Sprintf("%d B", size)
	}
}

func userHomeDirectory() (string, error) {
	home, err := config.DefaultPreference()
	if err != nil {
		return "", fmt.Errorf("get default local directory: %w", err)
	}

	if home.StartingLocalDirectory == "" {
		return "", errors.New("starting local directory is not configured")
	}

	return home.StartingLocalDirectory, nil
}