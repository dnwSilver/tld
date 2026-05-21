package app

import tea "github.com/charmbracelet/bubbletea"

func Run() error {
	_, err := tea.NewProgram(newModel(), tea.WithAltScreen()).Run()
	return err
}
