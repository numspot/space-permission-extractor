package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ProgressModel struct {
	logs   []LogMsg
	err    string
	width  int
	height int
}

func NewProgressModel() ProgressModel {
	return ProgressModel{}
}

func (m ProgressModel) Init() tea.Cmd {
	return nil
}

func (m ProgressModel) Update(msg tea.Msg) (ProgressModel, tea.Cmd) {
	return m, nil
}

func (m ProgressModel) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#89b4fa"))

	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#89b4fa"))
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	var lines []string
	lines = append(lines, titleStyle.Render("Extracting Permissions..."))
	lines = append(lines, "")

	if len(m.logs) == 0 {
		lines = append(lines, normalStyle.Render("Starting..."))
	} else {
		maxLogs := m.height - 6
		if maxLogs < 3 {
			maxLogs = 3
		}
		start := 0
		if len(m.logs) > maxLogs {
			start = len(m.logs) - maxLogs
		}

		for i := start; i < len(m.logs); i++ {
			log := m.logs[i]
			var style lipgloss.Style
			prefix := ""
			switch log.Level {
			case "info":
				style = infoStyle
				prefix = "* "
			case "warn":
				style = warnStyle
				prefix = "! "
			case "error":
				style = errorStyle
				prefix = "X "
			case "success":
				style = successStyle
				prefix = "+ "
			default:
				style = normalStyle
				prefix = ". "
			}

			msg := prefix + log.Message
			if len(msg) > m.width-4 {
				msg = msg[:m.width-7] + "..."
			}
			lines = append(lines, style.Render(msg))
		}
	}

	if m.err != "" {
		lines = append(lines, "")
		lines = append(lines, errorStyle.Render("Error: "+m.err))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m *ProgressModel) SetLogs(logs []LogMsg) {
	m.logs = logs
}

func (m *ProgressModel) SetError(err string) {
	m.err = err
}

func (m *ProgressModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}
