package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var titleStyle = lipgloss.NewStyle().Bold(true)
var textStyleCentered = lipgloss.NewStyle().Align(lipgloss.Center).Padding(2)
var textStyle = lipgloss.NewStyle().Padding(2)
var spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
var listHeaderStyle = lipgloss.NewStyle().BorderBottom(true).BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("69"))
var listItemStyle = lipgloss.NewStyle()

var xmrCoinArt = `   __
 /"  "\
|_|\/|_|
\      /
 "----"
`

func IdleView(m model) string {
	//return
	text := titleStyle.Render("Welcome to MoneroKon 2023") + "\n\n" +
		listHeaderStyle.Render("Current rate:") + "\n" +
		listItemStyle.Render(fmt.Sprintf("1 XMR = %.3f EUR", m.xmrPrice))

	w, h := lipgloss.Size(text)
	coin := lipgloss.NewStyle().PaddingLeft((m.width - w - 9) / 2).Render(xmrCoinArt)
	body := textStyle.Render(lipgloss.JoinHorizontal(0, text, coin))

	h = lipgloss.Height(body)
	tBlock := lipgloss.Place(m.width, m.height-h, lipgloss.Center, lipgloss.Bottom,
		titleStyle.Render("Touch or scan to begin")+"\n\nPowered by digilol.net\n\n")
	return lipgloss.JoinVertical(lipgloss.Left, body, tBlock)
	//return fmt.Sprintf("Displaying cool ads and animations. Press any key to start buying Monero.")
}

func AddressInView(m model) string {
	textBlock := textStyleCentered.Render(
		"Enter the receiving address or scan QR code:\n\n",
		m.textinput.View())
	h := lipgloss.Height(textBlock)
	tBlock := lipgloss.Place(m.width, m.height-h, lipgloss.Right, lipgloss.Bottom, m.timer.View())
	return lipgloss.JoinVertical(lipgloss.Right, textBlock, tBlock)
}

func MoneyInView(m model) string {
	textBlock := textStyleCentered.Render(m.spinner.View(),
		fmt.Sprintf("Received: %.2f EUR", float64(m.euro)/100),
		"\n\n", "Press enter to proceed.")

	h := lipgloss.Height(textBlock)
	tBlock := lipgloss.Place(m.width, m.height-h, lipgloss.Right, lipgloss.Bottom, m.timer.View())
	return lipgloss.JoinVertical(lipgloss.Right, textBlock, tBlock)
}

func TxInfoView(m model) string {
	return textStyleCentered.Render(fmt.Sprintf("TxId: %s\nAmount: %f\nFee: %f\nAddress: %s",
		"78b5e0c836fabc8d210f00a94f0e2da45c5d0a14cbba1baf47cd3137c632c3ff",
		float64(m.euro)/100, 0.0002, m.address))
}
