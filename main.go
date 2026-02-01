package main

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	saveFile := ""
	if len(os.Args) > 1 {
		saveFile = os.Args[1]
	}

	p := tea.NewProgram(NewModel(saveFile), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
