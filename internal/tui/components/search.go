package components

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SearchModel represents a search input component.
type SearchModel struct {
	Input  textinput.Model
	Active bool
}

// NewSearchModel creates a new search model.
func NewSearchModel() SearchModel {
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.Width = 40
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#f9e2af"))
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#cdd6f4"))
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#6c7086"))
	return SearchModel{
		Input:  ti,
		Active: false,
	}
}

// Activate activates the search input.
func (s *SearchModel) Activate() {
	s.Active = true
	s.Input.SetValue("")
	s.Input.Focus()
}

// Deactivate deactivates the search input.
func (s *SearchModel) Deactivate() {
	s.Active = false
	s.Input.Blur()
}

// Update updates the search model.
func (s SearchModel) Update(msg tea.Msg) (SearchModel, tea.Cmd) {
	if !s.Active {
		return s, nil
	}
	var cmd tea.Cmd
	s.Input, cmd = s.Input.Update(msg)
	return s, cmd
}

// SetWidth sets the width of the search input.
func (s *SearchModel) SetWidth(width int) {
	s.Input.Width = width - 20
	if s.Input.Width < 20 {
		s.Input.Width = 20
	}
}

// View renders the search input.
func (s SearchModel) View() string {
	if !s.Active {
		return ""
	}

	searchStyle := lipgloss.NewStyle().Padding(0, 1)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#f9e2af")).Bold(true)
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6c7086"))

	return searchStyle.Render(
		labelStyle.Render("Search: ") + s.Input.View() + hintStyle.Render("  [Esc to close | Ctrl+R to clear]"),
	)
}
