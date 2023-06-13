package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// TODO: Pimp my idle view
func IdleView(m model) string {
	return titleStyle.Render("Welcome to MoneroKon 2023")
	//return fmt.Sprintf("Displaying cool ads and animations. Press any key to start buying Monero.")
}

func AddressInView(m model) string {
	textBlock := textStyle.Render(
		"Enter the receiving address or scan QR code:\n\n",
		m.textinput.View())
	//tBlock := m.timerstyle.Render(m.timer.View())
	h := lipgloss.Height(textBlock)
	tBlock := lipgloss.Place(m.width, m.height-h, lipgloss.Right, lipgloss.Bottom, m.timer.View())
	return lipgloss.JoinVertical(lipgloss.Right, textBlock, tBlock)
}

func MoneyInView(m model) string {
	textBlock := textStyle.Render(m.spinner.View(),
		fmt.Sprintf("Received: %.2f EUR", float64(m.euro)),
		"\n\n", "Press enter to proceed.")

	h := lipgloss.Height(textBlock)
	tBlock := lipgloss.Place(m.width, m.height-h, lipgloss.Right, lipgloss.Bottom, m.timer.View())
	return lipgloss.JoinVertical(lipgloss.Right, textBlock, tBlock)
}

func TxInfoView(m model) string {
	return textStyle.Render(fmt.Sprintf("TxId: %s\nAmount: %s\nFee: %f\nAddress: %s", "blahblah", float64(m.euro), 0.0002, m.address))
}
