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
	allRows []api.CSVRow // unfiltered original rows
	mode    Mode
	scrollY int
	width   int
	height  int
	search  SearchModel
}

func NewResultsModel() ResultsModel {
	return ResultsModel{
		search: NewSearchModel(),
	}
}

func (m ResultsModel) Init() tea.Cmd {
	return nil
}

func (m ResultsModel) Update(msg tea.Msg) (ResultsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.search.Active {
			switch msg.String() {
			case "esc":
				m.search.Deactivate()
				m.rows = m.allRows
				m.scrollY = 0
				return m, nil
			case "ctrl+r":
				m.search.Input.SetValue("")
				m.rows = m.allRows
				m.scrollY = 0
				return m, nil
			case "up", "k":
				if m.scrollY > 0 {
					m.scrollY--
				}
				return m, nil
			case "down", "j":
				maxScroll := len(m.rows) - m.maxVisibleRows()
				if maxScroll < 0 {
					maxScroll = 0
				}
				if m.scrollY < maxScroll {
					m.scrollY++
				}
				return m, nil
			}
			oldValue := m.search.Input.Value()
			var cmd tea.Cmd
			m.search, cmd = m.search.Update(msg)
			if m.search.Input.Value() != oldValue {
				m.applySearch()
			}
			return m, cmd
		}

		switch msg.String() {
		case "/":
			m.search.Activate()
			m.search.SetWidth(m.width)
			return m, nil
		case "up", "k":
			if m.scrollY > 0 {
				m.scrollY--
			}
		case "down", "j":
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
	searchHeight := 0
	if m.search.Active {
		searchHeight = 1
	}
	if m.height < 8+searchHeight {
		return 1
	}
	return m.height - 6 - searchHeight
}

func (m ResultsModel) View() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#89b4fa"))
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#f9e2af"))
	cellStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	fadedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	var lines []string
	lines = append(lines, titleStyle.Render(fmt.Sprintf("Results: %d rows (%s)", len(m.rows), m.mode.DisplayName())))

	if m.search.Active {
		lines = append(lines, m.search.View())
	}

	if len(m.rows) == 0 {
		if len(m.allRows) > 0 {
			lines = append(lines, fadedStyle.Render("No matches"))
		} else {
			lines = append(lines, fadedStyle.Render("No results"))
		}
		lines = append(lines, fadedStyle.Render("Enter: export | /: search | Esc: back"))
		return lipgloss.JoinVertical(lipgloss.Left, lines...)
	}

	switch m.mode {
	case ModeResource, ModeIdentity:
		lines = append(lines, m.renderResourceIdentityTable(headerStyle, cellStyle, fadedStyle)...)
	default:
		lines = append(lines, m.renderUserSATable(headerStyle, cellStyle, fadedStyle)...)
	}

	lines = append(lines, fadedStyle.Render("Up/Down: scroll | Enter: export | /: search | Esc: back"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m *ResultsModel) applySearch() {
	query := strings.ToLower(m.search.Input.Value())
	if query == "" {
		m.rows = m.allRows
		return
	}
	m.scrollY = 0
	var filtered []api.CSVRow
	for _, row := range m.allRows {
		if m.rowMatches(row, query) {
			filtered = append(filtered, row)
		}
	}
	m.rows = filtered
}

func (m ResultsModel) rowMatches(row api.CSVRow, query string) bool {
	fields := []string{
		row.EntityID, row.EntityName, row.EntityEmail,
		row.ItemType, row.ItemName, row.ItemDescription,
		row.ResourceType, row.Domain, row.ItemAction,
	}
	for _, f := range fields {
		if strings.Contains(strings.ToLower(f), query) {
			return true
		}
	}
	return false
}

func (m ResultsModel) IsSearchActive() bool {
	return m.search.Active
}

type colDef struct {
	name   string
	field  func(api.CSVRow) string
	min    int
	weight float64
}

func distributeWidths(defs []colDef, available int, gap int) []int {
	n := len(defs)
	if n == 0 || available <= 0 {
		return make([]int, n)
	}

	minSum := 0
	weightSum := 0.0
	for _, d := range defs {
		minSum += d.min
		weightSum += d.weight
	}

	gaps := gap * (n - 1)
	if gaps < 0 {
		gaps = 0
	}

	extra := available - minSum - gaps
	if extra < 0 {
		extra = 0
	}

	widths := make([]int, n)
	for i, d := range defs {
		bonus := int(float64(extra) * (d.weight / weightSum))
		widths[i] = d.min + bonus
	}

	extraUsed := gaps
	for _, w := range widths {
		extraUsed += w
	}

	if diff := available - extraUsed; diff > 0 && weightSum > 0 {
		for i := range widths {
			if defs[i].weight > 0 {
				widths[i] += diff
				break
			}
		}
	}

	return widths
}

func (m ResultsModel) renderUserSATable(headerStyle, cellStyle, fadedStyle lipgloss.Style) []string {
	var lines []string

	gap := 1
	defs := []colDef{
		{name: "Entity", field: func(r api.CSVRow) string { return r.EntityName }, min: 8, weight: 3},
		{name: "ID", field: func(r api.CSVRow) string { return r.EntityID }, min: 12, weight: 3},
		{name: "Type", field: func(r api.CSVRow) string { return r.ItemType }, min: 6, weight: 1},
		{name: "Name", field: func(r api.CSVRow) string { return r.ItemName }, min: 8, weight: 2},
		{name: "Action", field: func(r api.CSVRow) string { return r.ItemAction }, min: 6, weight: 1},
		{name: "Description", field: func(r api.CSVRow) string { return r.ItemDescription }, min: 10, weight: 3},
	}

	widths := distributeWidths(defs, m.width, gap)
	totalW := 0
	for i, w := range widths {
		if i > 0 {
			totalW += gap
		}
		totalW += w
	}

	sep := strings.Repeat("─", totalW)

	var headerCells []string
	for i, d := range defs {
		headerCells = append(headerCells, fmt.Sprintf("%-*s", widths[i], truncate(d.name, widths[i])))
	}
	lines = append(lines, headerStyle.Render(strings.Join(headerCells, strings.Repeat(" ", gap))))
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
		var cells []string
		for j, d := range defs {
			val := d.field(row)
			cells = append(cells, fmt.Sprintf("%-*s", widths[j], truncate(val, widths[j])))
		}
		lines = append(lines, cellStyle.Render(strings.Join(cells, strings.Repeat(" ", gap))))
	}

	lines = append(lines, fadedStyle.Render(fmt.Sprintf("--- %d-%d/%d ---", start+1, end, len(m.rows))))

	return lines
}

func (m ResultsModel) renderResourceIdentityTable(headerStyle, cellStyle, fadedStyle lipgloss.Style) []string {
	var lines []string

	gap := 1
	defs := []colDef{
		{name: "ID", field: func(r api.CSVRow) string { return r.EntityID }, min: 8, weight: 2},
		{name: "Entity", field: func(r api.CSVRow) string { return r.EntityEmail }, min: 8, weight: 2},
		{name: "Type", field: func(r api.CSVRow) string { return r.EntityName }, min: 5, weight: 1},
		{name: "Domain", field: func(r api.CSVRow) string { return r.Domain }, min: 6, weight: 1},
		{name: "Resource", field: func(r api.CSVRow) string { return r.ResourceType }, min: 8, weight: 1},
		{name: "Auth", field: func(r api.CSVRow) string { return r.ItemType }, min: 6, weight: 1},
		{name: "ResourceID", field: func(r api.CSVRow) string { return r.ItemDescription }, min: 8, weight: 2},
		{name: "Name", field: func(r api.CSVRow) string { return r.ItemName }, min: 8, weight: 2},
	}

	widths := distributeWidths(defs, m.width, gap)
	totalW := 0
	for i, w := range widths {
		if i > 0 {
			totalW += gap
		}
		totalW += w
	}

	sep := strings.Repeat("─", totalW)

	var headerCells []string
	for i, d := range defs {
		headerCells = append(headerCells, fmt.Sprintf("%-*s", widths[i], truncate(d.name, widths[i])))
	}
	lines = append(lines, headerStyle.Render(strings.Join(headerCells, strings.Repeat(" ", gap))))
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
		var cells []string
		for j, d := range defs {
			val := d.field(row)
			cells = append(cells, fmt.Sprintf("%-*s", widths[j], truncate(val, widths[j])))
		}
		lines = append(lines, cellStyle.Render(strings.Join(cells, strings.Repeat(" ", gap))))
	}

	lines = append(lines, fadedStyle.Render(fmt.Sprintf("--- %d-%d/%d ---", start+1, end, len(m.rows))))

	return lines
}

func (m *ResultsModel) SetRows(rows []api.CSVRow, mode Mode) {
	// For ModeResource, filter out rows without an email address
	if mode == ModeResource {
		var filtered []api.CSVRow
		for _, r := range rows {
			if strings.TrimSpace(r.EntityEmail) != "" {
				filtered = append(filtered, r)
			}
		}
		rows = filtered
	}
	m.allRows = rows
	m.rows = rows
	m.mode = mode
	m.scrollY = 0
	if m.search.Active {
		m.search.Deactivate()
	}
}

func (m *ResultsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.search.SetWidth(width)
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
