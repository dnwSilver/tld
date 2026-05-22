package app

import tea "github.com/charmbracelet/bubbletea"

func isCancelKey(msg tea.KeyMsg) bool {
	return msg.Type == tea.KeyEsc || msg.String() == "esc"
}

func isQuitKey(msg tea.KeyMsg) bool {
	return msg.Type == tea.KeyCtrlC || msg.String() == "ctrl+c"
}

func isEnterKey(msg tea.KeyMsg) bool {
	return msg.Type == tea.KeyEnter || msg.String() == "enter"
}

func isBackspaceKey(msg tea.KeyMsg) bool {
	return msg.Type == tea.KeyBackspace || msg.String() == "backspace"
}

func isOneOf(msg tea.KeyMsg, values ...string) bool {
	key := msg.String()
	for _, value := range values {
		if key == value {
			return true
		}
	}

	return false
}
