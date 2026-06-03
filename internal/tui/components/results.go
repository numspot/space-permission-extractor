package components

import (
	"fmt"
	"strings"

	"permission-extractor-tui/internal/api"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ResultsModel struct {
	rows    []api.CSVRow
	mode    Mode
	scrollY int
	width   int
	height  int
}

func NewResultsModel() ResultsModel {
	return ResultsModel{}
}

func (m ResultsModel) Init() tea.Cmd {
	return nil
}

func (m ResultsModel) Update(msg tea.Msg) (ResultsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if m.scrollY > 0 {
				m.scrollY--
			}
		case "down":
			maxScroll := len(m.rows) - m.maxVisibleRows()
			if maxScroll < 0 {
				maxScroll = 0
			}
			if m.scrollY < maxScroll {
				m.scrollY++
			}
		}
	}

	return m, nil
}

func (m ResultsModel) maxVisibleRows() int {
	if m.height < 8 {
		return 1
	}
	return m.height - 6
}

func (m ResultsModel) View() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#89b4fa"))
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#f9e2af"))
	cellStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	fadedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	var lines []string
	lines = append(lines, titleStyle.Render(fmt.Sprintf("Results: %d rows (%s)", len(m.rows), m.mode.DisplayName())))

	if len(m.rows) == 0 {
		lines = append(lines, fadedStyle.Render("No results"))
		lines = append(lines, fadedStyle.Render("Enter: export | Esc: back"))
		return lipgloss.JoinVertical(lipgloss.Left, lines...)
	}

	switch m.mode {
	case ModeResource, ModeIdentity:
		lines = append(lines, m.renderResourceIdentityTable(headerStyle, cellStyle, fadedStyle)...)
	default:
		lines = append(lines, m.renderUserSATable(headerStyle, cellStyle, fadedStyle)...)
	}

	lines = append(lines, fadedStyle.Render("Up/Down: scroll | Enter: export | Esc: back"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m ResultsModel) renderUserSATable(headerStyle, cellStyle, fadedStyle lipgloss.Style) []string {
	var lines []string

	entityW := 40
	idW := 38
	typeW := 11
	nameW := 25
	actionW := 15
	separatorCount := 5
	descW := m.width - entityW - idW - typeW - nameW - actionW - separatorCount
	if descW < 15 {
		descW = 15
	}
	totalW := entityW + idW + typeW + nameW + actionW + descW + separatorCount

	sep := strings.Repeat("─", totalW)

	lines = append(lines, headerStyle.Render(fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s %-*s", entityW, "Entity", idW, "ID", typeW, "Type", nameW, "Name", actionW, "Action", descW, "Description")))
	lines = append(lines, fadedStyle.Render(sep))

	maxRows := m.maxVisibleRows()
	start := m.scrollY
	if start < 0 {
		start = 0
	}
	end := start + maxRows
	if end > len(m.rows) {
		end = len(m.rows)
	}

	for i := start; i < end; i++ {
		row := m.rows[i]
		entity := truncate(row.EntityName, entityW-1)
		id := truncate(row.EntityID, idW-1)
		itemType := truncate(row.ItemType, typeW-1)
		name := truncate(row.ItemName, nameW-1)
		action := truncate(row.ItemAction, actionW-1)
		desc := truncate(row.ItemDescription, descW-1)
		lines = append(lines, cellStyle.Render(fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s %-*s", entityW, entity, idW, id, typeW, itemType, nameW, name, actionW, action, descW, desc)))
	}

	lines = append(lines, fadedStyle.Render(fmt.Sprintf("--- %d-%d/%d ---", start+1, end, len(m.rows))))

	return lines
}

func (m ResultsModel) renderResourceIdentityTable(headerStyle, cellStyle, fadedStyle lipgloss.Style) []string {
	var lines []string

	idW := 38
	entityW := 30
	typeW := 8
	domainW := 12
	resourceTypeW := 14
	authTypeW := 12
	resourceIdW := 38
	separatorCount := 7
	nameW := m.width - idW - entityW - typeW - domainW - resourceTypeW - authTypeW - resourceIdW - separatorCount
	if nameW < 15 {
		nameW = 15
	}
	totalW := idW + entityW + typeW + domainW + resourceTypeW + authTypeW + resourceIdW + nameW + separatorCount

	sep := strings.Repeat("─", totalW)

	lines = append(lines, headerStyle.Render(fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s", idW, "ID", entityW, "Entity", typeW, "Type", domainW, "Domain", resourceTypeW, "Resource", authTypeW, "Auth", resourceIdW, "ResourceID", nameW, "Name")))
	lines = append(lines, fadedStyle.Render(sep))

	maxRows := m.maxVisibleRows()
	start := m.scrollY
	if start < 0 {
		start = 0
	}
	end := start + maxRows
	if end > len(m.rows) {
		end = len(m.rows)
	}

	for i := start; i < end; i++ {
		row := m.rows[i]
		id := truncate(row.EntityID, idW-1)
		entity := truncate(row.EntityEmail, entityW-1)
		itemType := truncate(row.EntityName, typeW-1)
		domain := truncate(row.Domain, domainW-1)
		resourceType := truncate(row.ResourceType, resourceTypeW-1)
		authType := truncate(row.ItemType, authTypeW-1)
		resourceId := truncate(row.ItemDescription, resourceIdW-1)
		name := truncate(row.ItemName, nameW-1)
		lines = append(lines, cellStyle.Render(fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s", idW, id, entityW, entity, typeW, itemType, domainW, domain, resourceTypeW, resourceType, authTypeW, authType, resourceIdW, resourceId, nameW, name)))
	}

	lines = append(lines, fadedStyle.Render(fmt.Sprintf("--- %d-%d/%d ---", start+1, end, len(m.rows))))

	return lines
}

func (m *ResultsModel) SetRows(rows []api.CSVRow, mode Mode) {
	m.rows = rows
	m.mode = mode
	m.scrollY = 0
}

func (m *ResultsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max < 2 {
		return s[:max]
	}
	return s[:max-2] + ".."
}
