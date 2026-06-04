package tui

import (
	"context"
	"strings"

	"permission-extractor-tui/internal/api"
	"permission-extractor-tui/internal/config"
	"permission-extractor-tui/internal/tui/components"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type App struct {
	config      config.Config
	screen      Screen
	mode        components.Mode
	resourceID  string
	identityID  string
	rows        []api.CSVRow
	err         error
	width       int
	height      int
	logMessages []LogMsg

	configModel   components.ConfigModel
	modeModel     components.ModeModel
	progressModel components.ProgressModel
	resultsModel  components.ResultsModel
	exportModel   components.ExportModel

	ctx    context.Context
	cancel context.CancelFunc
}

func NewApp(cfg config.Config) *App {
	return &App{
		config:        cfg,
		screen:        ScreenConfig,
		mode:          components.ModeUser,
		configModel:   components.NewConfigModel(cfg),
		modeModel:     components.NewModeModel(),
		progressModel: components.NewProgressModel(),
		resultsModel:  components.NewResultsModel(),
		exportModel:   components.NewExportModel(),
	}
}

func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.configModel.Init(),
		a.modeModel.Init(),
	)
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// When search is active in Results, only allow ctrl+c globally
		if a.screen == ScreenResults && a.resultsModel.IsSearchActive() {
			if msg.String() == "ctrl+c" {
				return a, tea.Quit
			}
			break
		}

		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit
		case "esc":
			if a.screen == ScreenProgress {
				a.cancel()
				a.progressModel = components.NewProgressModel()
				a.logMessages = nil
				a.err = nil
				a.screen = ScreenMode
				return a, nil
			}
			if a.screen == ScreenConfig {
				return a, tea.Quit
			}
			if a.screen == ScreenExport {
				a.screen = ScreenResults
				return a, nil
			}
			if a.screen == ScreenResults {
				a.screen = ScreenMode
				a.modeModel.Reset()
				return a, nil
			}
			if a.screen == ScreenMode {
				if a.modeModel.GetPhase() == components.PhaseSelect {
					a.screen = ScreenConfig
					return a, nil
				}
			}
		case "ctrl+s":
			if a.screen != ScreenConfig && a.screen != ScreenProgress {
				if a.screen == ScreenMode && a.modeModel.IsInputFocused() {
				} else {
					a.screen = ScreenConfig
					return a, nil
				}
			}

		case "enter":
			switch a.screen {
			case ScreenConfig:
				if a.configModel.GetFocusIndex() == 4 {
					a.configModel.SaveConfig()
				} else if a.configModel.Valid() {
					a.config = a.configModel.GetConfig()
					a.screen = ScreenMode
				}
			case ScreenMode:
				if a.modeModel.CanStart() {
					a.mode = a.modeModel.GetMode()
					a.resourceID = a.modeModel.GetResourceID()
					a.identityID = a.modeModel.GetIdentityID()
					a.screen = ScreenProgress
					a.logMessages = nil
					a.ctx, a.cancel = context.WithCancel(context.Background())
					cmds = append(cmds, a.startExtraction())
				}
			case ScreenResults:
				a.exportModel.SetRows(a.rows, a.mode)
				a.screen = ScreenExport
			case ScreenExport:
				if err := a.exportModel.Export(); err == nil {
					a.screen = ScreenResults
				}
			}
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.configModel.SetSize(msg.Width-8, msg.Height-14)
		a.modeModel.SetSize(msg.Width-8, msg.Height-14)
		a.progressModel.SetSize(msg.Width-8, msg.Height-14)
		a.resultsModel.SetSize(msg.Width-8, msg.Height-14)
		a.exportModel.SetSize(msg.Width-8, msg.Height-14)

	case LogMsg:
		a.logMessages = append(a.logMessages, msg)
		a.progressModel.SetLogs(convLogs(a.logMessages))

	case ExtractionCompleteMsg:
		a.rows = msg.Rows
		a.err = msg.Err
		if msg.Err != nil {
			if msg.Err == context.Canceled {
				a.screen = ScreenMode
			} else {
				a.progressModel.SetError(msg.Err.Error())
			}
		} else {
			a.screen = ScreenResults
			a.resultsModel.SetRows(msg.Rows, a.mode)
		}
	}

	switch a.screen {
	case ScreenConfig:
		var cmd tea.Cmd
		a.configModel, cmd = a.configModel.Update(msg)
		cmds = append(cmds, cmd)
	case ScreenMode:
		var cmd tea.Cmd
		a.modeModel, cmd = a.modeModel.Update(msg)
		cmds = append(cmds, cmd)
	case ScreenResults:
		var cmd tea.Cmd
		a.resultsModel, cmd = a.resultsModel.Update(msg)
		cmds = append(cmds, cmd)
	case ScreenExport:
		var cmd tea.Cmd
		a.exportModel, cmd = a.exportModel.Update(msg)
		cmds = append(cmds, cmd)
	}

	return a, tea.Batch(cmds...)
}

func (a *App) View() string {
	if a.width == 0 || a.height == 0 {
		return "Loading..."
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#f9e2af"))

	logo := " _  _                         _   \n" +
		"| \\| |_  _ _ __  ____ __  ___| |_ \n" +
		"| .` | || | '  \\(_-< '_ \\/ _ \\  _|\n" +
		"|_|\\_|\\_,_|_|_|_/__/ .__/\\___/\\__|\n" +
		"                   |_|\n" +
		"    -- Permission Extractor --"
	title := titleStyle.Render(logo)

	sidebarStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#f9e2af"))

	var content string
	switch a.screen {
	case ScreenConfig:
		content = a.configModel.View()
	case ScreenMode:
		content = a.modeModel.View()
	case ScreenProgress:
		content = a.progressModel.View()
	case ScreenResults:
		content = a.resultsModel.View()
	case ScreenExport:
		content = a.exportModel.View()
	}

	screens := []string{"Config", "Mode", "Progress", "Results", "Export"}
	var navItems []string
	for i, s := range screens {
		if Screen(i) == a.screen {
			navItems = append(navItems, activeStyle.Render("> "+s))
		} else {
			navItems = append(navItems, sidebarStyle.Render("  "+s))
		}
	}
	nav := strings.Join(navItems, "  ")

	contentWidth := a.width - 6
	if contentWidth < 30 {
		contentWidth = 30
	}
	mainHeight := a.height - 11
	if mainHeight < 6 {
		mainHeight = 6
	}

	contentBox := lipgloss.NewStyle().
		Width(contentWidth).
		Height(mainHeight).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("250")).
		Padding(1, 1).
		Render(content)

	mainContent := lipgloss.JoinVertical(lipgloss.Left, nav, contentBox)

	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	footer := footerStyle.Render(a.getHelpText())

	return lipgloss.JoinVertical(lipgloss.Left, title, "", mainContent, footer)
}

func (a *App) getHelpText() string {
	switch a.screen {
	case ScreenConfig:
		return "Tab: next | Shift+Tab: prev | Enter: continue | Esc: quit"
	case ScreenMode:
		return "Up/Down: select | Enter: confirm | Ctrl+S: settings | Esc: back"
	case ScreenProgress:
		return "Esc: cancel"
	case ScreenResults:
		return "Up/Down: scroll | Enter: export | /: search | Ctrl+S: settings | Esc: back"
	case ScreenExport:
		return "Enter: save | Ctrl+S: settings | Esc: back"
	default:
		return ""
	}
}

func (a *App) startExtraction() tea.Cmd {
	return func() tea.Msg {
		client := api.NewClient(a.config)

		logChan := make(chan api.LogMessage, 100)
		go func() {
			for logMsg := range logChan {
				a.logMessages = append(a.logMessages, LogMsg{
					Level:   logMsg.Level,
					Message: logMsg.Message,
				})
			}
		}()

		var rows []api.CSVRow
		var err error

		switch a.mode {
		case components.ModeUser:
			rows, err = client.ProcessUsers(a.ctx, logChan)
		case components.ModeServiceAccount:
			rows, err = client.ProcessServiceAccounts(a.ctx, logChan)
		case components.ModeResource:
			rows, err = client.ProcessByResource(a.ctx, a.resourceID, logChan)
		case components.ModeIdentity:
			rows, err = client.ProcessByIdentity(a.ctx, a.identityID, logChan)
		}

		close(logChan)
		return ExtractionCompleteMsg{Rows: rows, Err: err}
	}
}
