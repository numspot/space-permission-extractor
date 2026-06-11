package components

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"permission-extractor-tui/internal/api"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ExportModel struct {
	rows     []api.CSVRow
	mode     Mode
	filename string
	exported bool
	err      string
	width    int
	height   int
}

func NewExportModel() ExportModel {
	home, _ := os.UserHomeDir()
	downloads := filepath.Join(home, "Downloads")
	_ = os.MkdirAll(downloads, 0755)
	return ExportModel{
		filename: filepath.Join(downloads, fmt.Sprintf("permissions_%s.csv", time.Now().Format("20060102_150405"))),
	}
}

func (m ExportModel) Init() tea.Cmd {
	return nil
}

func (m ExportModel) Update(msg tea.Msg) (ExportModel, tea.Cmd) {
	return m, nil
}

func (m ExportModel) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#89b4fa"))

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

	var lines []string
	lines = append(lines, titleStyle.Render("Export Results"))
	lines = append(lines, "")
	lines = append(lines, labelStyle.Render(fmt.Sprintf("Rows: %d", len(m.rows))))
	lines = append(lines, labelStyle.Render(fmt.Sprintf("Mode: %s", m.mode.DisplayName())))
	lines = append(lines, "")
	lines = append(lines, labelStyle.Render("Output file:"))
	lines = append(lines, "  "+filepath.Base(m.filename))
	lines = append(lines, "")

	if m.exported && m.err == "" {
		lines = append(lines, successStyle.Render("File saved successfully!"))
	} else if m.err != "" {
		lines = append(lines, errorStyle.Render("Error: "+m.err))
	}

	lines = append(lines, "")
	lines = append(lines, labelStyle.Render("Press Enter to save | q: back"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m *ExportModel) SetRows(rows []api.CSVRow, mode Mode) {
	m.rows = rows
	m.mode = mode
	m.exported = false
	m.err = ""
	home, _ := os.UserHomeDir()
	downloads := filepath.Join(home, "Downloads")
	_ = os.MkdirAll(downloads, 0755)
	switch mode {
	case ModeResource:
		m.filename = filepath.Join(downloads, fmt.Sprintf("resource_access_%s.csv", time.Now().Format("20060102_150405")))
	case ModeIdentity:
		m.filename = filepath.Join(downloads, fmt.Sprintf("identity_resources_%s.csv", time.Now().Format("20060102_150405")))
	case ModeServiceAccount:
		m.filename = filepath.Join(downloads, fmt.Sprintf("sa_access_%s.csv", time.Now().Format("20060102_150405")))
	default:
		m.filename = filepath.Join(downloads, fmt.Sprintf("user_access_%s.csv", time.Now().Format("20060102_150405")))
	}
}

func (m *ExportModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *ExportModel) Export() error {
	if len(m.rows) == 0 {
		m.err = "no data to export"
		return fmt.Errorf(m.err)
	}

	file, err := os.Create(m.filename)
	if err != nil {
		m.err = err.Error()
		return err
	}
	defer func() { _ = file.Close() }()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	var header []string
	switch m.mode {
	case ModeResource, ModeIdentity:
		header = []string{"User/SA UUID", "User/SA Email", "Identity Type", "Domain", "Resource Type", "Resource UUID", "Authorisation Type", "Authorisation UUID", "Authorisation Name", "Action"}
	default:
		header = []string{"Entity UUID", "Entity Name", "Resource UUID", "Authorisation Type", "Authorisation UUID", "Authorisation Name"}
	}

	if err := writer.Write(header); err != nil {
		m.err = err.Error()
		return err
	}

	for _, row := range api.SanitizeRows(m.rows) {
		var record []string
		switch m.mode {
		case ModeResource, ModeIdentity:
			record = []string{
				row.EntityID,
				row.EntityEmail,
				row.EntityName,
				row.Domain,
				row.ResourceType,
				row.ItemDescription,
				row.ItemType,
				row.ItemId,
				row.ItemName,
				row.ItemAction,
			}
		default:
			record = []string{
				row.EntityID,
				row.EntityName,
				row.ItemDescription,
				row.ItemType,
				row.ItemId,
				row.ItemName,
			}
		}
		if err := writer.Write(record); err != nil {
			m.err = err.Error()
			return err
		}
	}

	m.exported = true
	return nil
}
