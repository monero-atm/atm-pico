package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
)

func IdleView(m model) string {
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
}

func AddressInView(m model) string {
	textBlock := textStyleCentered.Render(
		"Enter the receiving address or scan QR code:\n\n",
		m.textinput.View())
	timerBlock := textStyleCentered.Render("Returning in", m.timer.View())

	okButton := activeButtonStyle.Render("Next")
	cancelButton := buttonStyle.Render("Cancel")

	buttons := lipgloss.JoinHorizontal(lipgloss.Top, zone.Mark("back", cancelButton), zone.Mark("next", okButton))

	return lipgloss.JoinVertical(lipgloss.Center, textBlock, timerBlock, buttons)
}

func MoneyInView(m model) string {
	textBlock := textStyleCentered.Render(m.spinner.View(),
		fmt.Sprintf("Received: %.2f EUR", float64(m.euro)/100))

	timerBlock := textStyleCentered.Render("Returning in", m.timer.View())

	okButton := activeButtonStyle.Render("Next")
	cancelButton := buttonStyle.Render("Cancel")

	buttons := lipgloss.JoinHorizontal(lipgloss.Top, zone.Mark("back", cancelButton), zone.Mark("next", okButton))

	return lipgloss.JoinVertical(lipgloss.Center, textBlock, timerBlock, buttons)
}

func TxInfoView(m model) string {
	return textStyleCentered.Render(fmt.Sprintf("TxId: %s\nAmount: %f\nFee: %f\nAddress: %s",
		"78b5e0c836fabc8d210f00a94f0e2da45c5d0a14cbba1baf47cd3137c632c3ff",
		float64(m.euro)/100, 0.0002, m.address))
}
