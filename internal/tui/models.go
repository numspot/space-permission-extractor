package tui

import (
	"permission-extractor-tui/internal/api"
	"permission-extractor-tui/internal/tui/components"
)

type Screen int

const (
	ScreenConfig Screen = iota
	ScreenMode
	ScreenProgress
	ScreenResults
	ScreenExport
)

type ExtractionCompleteMsg struct {
	Rows []api.CSVRow
	Err  error
}

type LogMsg struct {
	Level   string
	Message string
}

func convLogs(logs []LogMsg) []components.LogMsg {
	result := make([]components.LogMsg, len(logs))
	for i, l := range logs {
		result[i] = components.LogMsg{Level: l.Level, Message: l.Message}
	}
	return result
}
