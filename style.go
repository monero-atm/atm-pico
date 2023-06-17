package main

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle        = lipgloss.NewStyle().Bold(true)
	textStyleCentered = lipgloss.NewStyle().Align(lipgloss.Center).Padding(2)
	textStyle         = lipgloss.NewStyle().Padding(2)
	spinnerStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	listHeaderStyle   = lipgloss.NewStyle().BorderBottom(true).BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("69"))
	listItemStyle     = lipgloss.NewStyle()
	buttonStyle       = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFF7DB")).
				Background(lipgloss.Color("#F25D94")).
				Padding(1, 3).
				MarginRight(5)

	activeButtonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFF7DB")).
				Background(lipgloss.Color("69")).
				Padding(1, 3).
				MarginLeft(5).
				Bold(true)

	xmrCoinArt = `   __
 /"  "\
|_|\/|_|
\      /
 "----"
`
)
