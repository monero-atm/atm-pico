package main

import "github.com/charmbracelet/lipgloss"

var (
	green  = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
	subtle = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}

	divider = lipgloss.NewStyle().
		SetString("•").
		Padding(0, 1).
		Foreground(subtle).
		String()
	urlStyle  = lipgloss.NewStyle().Foreground(green)
	checkMark = lipgloss.NewStyle().SetString("✓").
			Foreground(green).
			PaddingRight(1).
			String()
	cross = lipgloss.NewStyle().SetString("x").
		Foreground(lipgloss.Color("2")).
		PaddingRight(1).
		String()

	titleStyle        = lipgloss.NewStyle().Bold(true)
	textStyleCentered = lipgloss.NewStyle().Align(lipgloss.Center).Padding(2)
	textStyle         = lipgloss.NewStyle().Padding(2)
	spinnerStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))

	purpleListHeaderStyle = lipgloss.NewStyle().BorderBottom(true).BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("69"))

	pinkListHeaderStyle = lipgloss.NewStyle().BorderBottom(true).BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("201"))

	orangeListHeaderStyle = lipgloss.NewStyle().BorderBottom(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("202"))

	listStyle   = lipgloss.NewStyle().MarginRight(3)
	buttonStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("201")).
			Padding(1, 3).
			MarginRight(5)

	activeButtonStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("69")).
				Padding(1, 3).
				MarginLeft(5).
				Bold(true)

	doneButtonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF7DB")).
			Background(lipgloss.Color("69")).
			Padding(1, 3).
			Bold(true)
)
