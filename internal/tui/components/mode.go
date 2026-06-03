package components

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Mode int

const (
	ModeUser Mode = iota
	ModeServiceAccount
	ModeResource
	ModeIdentity
)

func (m Mode) String() string {
	switch m {
	case ModeUser:
		return "user"
	case ModeServiceAccount:
		return "sa"
	case ModeResource:
		return "resource"
	case ModeIdentity:
		return "identity"
	default:
		return "unknown"
	}
}

func (m Mode) DisplayName() string {
	switch m {
	case ModeUser:
		return "Users"
	case ModeServiceAccount:
		return "Service Accounts"
	case ModeResource:
		return "By Resource"
	case ModeIdentity:
		return "By Identity"
	default:
		return "Unknown"
	}
}

type ModePhase int

const (
	PhaseSelect ModePhase = iota
	PhaseInputResource
	PhaseInputIdentity
)

type LogMsg struct {
	Level   string
	Message string
}

type ModeModel struct {
	modeIndex  int
	phase      ModePhase
	resourceID textinput.Model
	identityID textinput.Model
	width      int
	height     int
}

func NewModeModel() ModeModel {
	resourceID := textinput.New()
	resourceID.Placeholder = "Paste or type UUID"
	resourceID.Width = 50

	identityID := textinput.New()
	identityID.Placeholder = "Paste or type UUID"
	identityID.Width = 50

	return ModeModel{
		modeIndex:  0,
		phase:      PhaseSelect,
		resourceID: resourceID,
		identityID: identityID,
	}
}

func (m ModeModel) Init() tea.Cmd {
	return nil
}

func (m ModeModel) Update(msg tea.Msg) (ModeModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.phase {
		case PhaseSelect:
			switch msg.String() {
			case "up":
				m.modeIndex = (m.modeIndex - 1 + 4) % 4
			case "down":
				m.modeIndex = (m.modeIndex + 1) % 4
			case "enter":
				switch Mode(m.modeIndex) {
				case ModeResource:
					m.phase = PhaseInputResource
					m.resourceID.Focus()
				case ModeIdentity:
					m.phase = PhaseInputIdentity
					m.identityID.Focus()
				}
			}
		case PhaseInputResource:
			switch msg.String() {
			case "esc":
				m.phase = PhaseSelect
				m.resourceID.Blur()
				return m, nil
			case "enter":
				return m, nil
			}
			m.resourceID, cmd = m.resourceID.Update(msg)
			return m, cmd
		case PhaseInputIdentity:
			switch msg.String() {
			case "esc":
				m.phase = PhaseSelect
				m.identityID.Blur()
				return m, nil
			case "enter":
				return m, nil
			}
			m.identityID, cmd = m.identityID.Update(msg)
			return m, cmd
		}
	}

	return m, cmd
}

func (m ModeModel) View() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#89b4fa"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#89b4fa"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	fadedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	var lines []string

	switch m.phase {
	case PhaseSelect:
		lines = append(lines, titleStyle.Render("Select Extraction Mode"))
		lines = append(lines, "")

		modes := []Mode{ModeUser, ModeServiceAccount, ModeResource, ModeIdentity}
		descriptions := []string{
			"Extract roles, permissions, and ACLs for all users",
			"Extract roles, permissions, and ACLs for service accounts",
			"List all identities with access to a specific resource",
			"List all resources accessible by a specific identity (user)",
		}

		for i, mode := range modes {
			prefix := "  "
			if i == m.modeIndex {
				prefix = "> "
				lines = append(lines, selectedStyle.Render(prefix+mode.DisplayName()))
			} else {
				lines = append(lines, labelStyle.Render(prefix+mode.DisplayName()))
			}
		}

		lines = append(lines, "")
		lines = append(lines, labelStyle.Render("Description:"))
		lines = append(lines, "  "+descriptions[m.modeIndex])
		lines = append(lines, "")
		lines = append(lines, fadedStyle.Render("Up/Down: select | Enter: confirm | Esc: back"))

	case PhaseInputResource:
		lines = append(lines, titleStyle.Render("By Resource"))
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render("Enter Resource UUID:"))
		lines = append(lines, "")
		lines = append(lines, "  "+m.resourceID.View())
		lines = append(lines, "")
		lines = append(lines, fadedStyle.Render("Enter: start | Esc: back"))

	case PhaseInputIdentity:
		lines = append(lines, titleStyle.Render("By Identity"))
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render("Enter User UUID:"))
		lines = append(lines, "")
		lines = append(lines, "  "+m.identityID.View())
		lines = append(lines, "")
		lines = append(lines, fadedStyle.Render("Enter: start | Esc: back"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m ModeModel) GetMode() Mode {
	return Mode(m.modeIndex)
}

func (m ModeModel) GetResourceID() string {
	return m.resourceID.Value()
}

func (m ModeModel) GetIdentityID() string {
	return m.identityID.Value()
}

func (m ModeModel) GetPhase() ModePhase {
	return m.phase
}

func (m *ModeModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.resourceID.Width = width - 10
	m.identityID.Width = width - 10
}

func (m *ModeModel) Reset() {
	m.phase = PhaseSelect
	m.resourceID.SetValue("")
	m.identityID.SetValue("")
	m.resourceID.Blur()
	m.identityID.Blur()
}

func (m ModeModel) IsInputFocused() bool {
	return m.phase == PhaseInputResource || m.phase == PhaseInputIdentity
}

func (m ModeModel) CanStart() bool {
	switch Mode(m.modeIndex) {
	case ModeResource:
		return m.phase == PhaseInputResource && m.resourceID.Value() != ""
	case ModeIdentity:
		return m.phase == PhaseInputIdentity && m.identityID.Value() != ""
	default:
		return m.phase == PhaseSelect
	}
}
