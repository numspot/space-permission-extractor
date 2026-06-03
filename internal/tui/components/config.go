package components

import (
	"strings"

	"permission-extractor-tui/internal/config"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ConfigModel struct {
	clientID     textinput.Model
	clientSecret textinput.Model
	spaceID      textinput.Model
	baseURL      textinput.Model
	focusIndex   int
	saved        bool
	saveError    string
	width        int
	height       int
}

func NewConfigModel(cfg config.Config) ConfigModel {
	clientID := textinput.New()
	clientID.Placeholder = "OAuth2 Client ID"
	clientID.Width = 50
	clientID.SetValue(cfg.ClientID)

	clientSecret := textinput.New()
	clientSecret.Placeholder = "OAuth2 Client Secret"
	clientSecret.Width = 50
	clientSecret.EchoMode = textinput.EchoPassword
	clientSecret.EchoCharacter = '*'
	clientSecret.SetValue(cfg.ClientSecret)

	spaceID := textinput.New()
	spaceID.Placeholder = "Space UUID"
	spaceID.Width = 50
	spaceID.SetValue(cfg.SpaceID)

	baseURL := textinput.New()
	baseURL.Placeholder = "https://api.eu-west-2.numspot.com"
	baseURL.Width = 50
	baseURL.SetValue(cfg.BaseURL)

	m := ConfigModel{
		clientID:     clientID,
		clientSecret: clientSecret,
		spaceID:      spaceID,
		baseURL:      baseURL,
	}
	m.clientID.Focus()
	return m
}

func (m ConfigModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m ConfigModel) Update(msg tea.Msg) (ConfigModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.focusIndex = (m.focusIndex + 1) % 5
			m.updateFocus()
			return m, nil
		case "shift+tab":
			m.focusIndex = (m.focusIndex - 1 + 5) % 5
			m.updateFocus()
			return m, nil
		}
	}

	switch m.focusIndex {
	case 0:
		m.clientID, cmd = m.clientID.Update(msg)
	case 1:
		m.clientSecret, cmd = m.clientSecret.Update(msg)
	case 2:
		m.spaceID, cmd = m.spaceID.Update(msg)
	case 3:
		m.baseURL, cmd = m.baseURL.Update(msg)
	}

	return m, cmd
}

func (m *ConfigModel) updateFocus() {
	m.clientID.Blur()
	m.clientSecret.Blur()
	m.spaceID.Blur()
	m.baseURL.Blur()

	switch m.focusIndex {
	case 0:
		m.clientID.Focus()
	case 1:
		m.clientSecret.Focus()
	case 2:
		m.spaceID.Focus()
	case 3:
		m.baseURL.Focus()
	}
}

func (m ConfigModel) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#89b4fa"))

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	focusedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#89b4fa"))
	buttonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#89b4fa")).
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#89b4fa"))
	buttonInactiveStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("241"))
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

	var lines []string
	lines = append(lines, titleStyle.Render("Configuration"))
	lines = append(lines, "")

	labels := []string{"Client ID", "Client Secret", "Space ID", "Base URL"}
	inputs := []textinput.Model{m.clientID, m.clientSecret, m.spaceID, m.baseURL}

	for i, label := range labels {
		var style lipgloss.Style
		if i == m.focusIndex {
			style = focusedStyle
		} else {
			style = labelStyle
		}
		lines = append(lines, style.Render(label+":"))

		if i == m.focusIndex {
			lines = append(lines, "> "+inputs[i].View())
		} else {
			lines = append(lines, "  "+inputs[i].View())
		}
		lines = append(lines, "")
	}

	if m.focusIndex == 4 {
		lines = append(lines, buttonStyle.Render("[ Save to ./.env ]"))
	} else {
		lines = append(lines, buttonInactiveStyle.Render("[ Save to ./.env ]"))
	}

	if m.saved && m.saveError == "" {
		lines = append(lines, successStyle.Render("Saved!"))
	} else if m.saveError != "" {
		lines = append(lines, errorStyle.Render("Error: "+m.saveError))
	}

	lines = append(lines, "")

	valid := m.Valid()
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	if valid {
		statusStyle = statusStyle.Foreground(lipgloss.Color("#89b4fa"))
	}

	var status string
	if valid {
		status = "Ready - Press Enter to continue"
	} else {
		missing := []string{}
		if m.clientID.Value() == "" {
			missing = append(missing, "Client ID")
		}
		if m.clientSecret.Value() == "" {
			missing = append(missing, "Client Secret")
		}
		if m.spaceID.Value() == "" {
			missing = append(missing, "Space ID")
		}
		status = "Missing: " + strings.Join(missing, ", ")
	}

	lines = append(lines, statusStyle.Render(status))

	return strings.Join(lines, "\n")
}

func (m ConfigModel) Valid() bool {
	return m.clientID.Value() != "" &&
		m.clientSecret.Value() != "" &&
		m.spaceID.Value() != ""
}

func (m ConfigModel) GetConfig() config.Config {
	baseURL := m.baseURL.Value()
	if baseURL == "" {
		baseURL = "https://api.eu-west-2.numspot.com"
	}
	return config.Config{
		ClientID:     m.clientID.Value(),
		ClientSecret: m.clientSecret.Value(),
		SpaceID:      m.spaceID.Value(),
		BaseURL:      baseURL,
	}
}

func (m *ConfigModel) SaveConfig() error {
	cfg := m.GetConfig()
	err := cfg.SaveToEnvFile()
	if err != nil {
		m.saveError = err.Error()
		return err
	}
	m.saved = true
	m.saveError = ""
	return nil
}

func (m *ConfigModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.clientID.Width = width - 10
	m.clientSecret.Width = width - 10
	m.spaceID.Width = width - 10
	m.baseURL.Width = width - 10
}

func (m ConfigModel) GetFocusIndex() int {
	return m.focusIndex
}
