package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
)

func IdleView(m model) string {
	motd := titleStyle.Render(cfg.Motd) + "\n\n"
	spacing := (m.width - 4) / 3
	rate := listStyle.Width(spacing).Render(purpleListHeaderStyle.Render("Rate:") + "\n" +
		fmt.Sprintf("1 XMR = %.3f %s", m.xmrPrice,
			cfg.CurrencyShort))

	fee := listStyle.Width(spacing).Render(pinkListHeaderStyle.Render("ATM fee:") + "\n" +
		fmt.Sprintf("%.2f", cfg.Fee) + "%")

	status := listStyle.Width(spacing).Render(orangeListHeaderStyle.Render("Status:") + "\n" +
		"Connected " + checkMark)

	body := textStyle.Render(motd + lipgloss.JoinHorizontal(lipgloss.Top, rate, fee, status))

	h := lipgloss.Height(body)
	tBlock := lipgloss.Place(m.width, m.height-h, lipgloss.Center, lipgloss.Bottom,
		titleStyle.Render("Touch or scan to begin")+"\n\nPowered by Digilol"+
			divider+urlStyle.Render("www.digilol.net")+"\n\n")
	return lipgloss.JoinVertical(lipgloss.Left, body, tBlock)
}

func AddressInView(m model) string {
	errMsg := ""
	if m.err != nil {
		errMsg = m.err.Error()
	}
	textBlock := textStyleCentered.Render(
		"Enter the receiving address or scan QR code:\n\n",
		m.textinput.View(), "\n", errMsg)
	timerBlock := textStyleCentered.Render("Returning in", m.timer.View())

	okButton := activeButtonStyle.Render("Next")
	cancelButton := buttonStyle.Render("Cancel")

	buttons := lipgloss.JoinHorizontal(lipgloss.Top, zone.Mark("back", cancelButton), zone.Mark("next", okButton))

	return lipgloss.JoinVertical(lipgloss.Center, textBlock, timerBlock, buttons)
}

func MoneyInView(m model) string {
	textBlock := textStyleCentered.Render(m.spinner.View(),
		fmt.Sprintf("Received: %.2f %s", float64(m.fiat)/100, cfg.CurrencyShort))

	timerBlock := textStyleCentered.Render("Returning in", m.timer.View())

	okButton := activeButtonStyle.Render("Next")
	cancelButton := buttonStyle.Render("Cancel")

	buttons := lipgloss.JoinHorizontal(lipgloss.Top, zone.Mark("back", cancelButton), zone.Mark("next", okButton))

	return lipgloss.JoinVertical(lipgloss.Center, textBlock, timerBlock, buttons)
}

func TxInfoView(m model) string {
	textBlock := ""
	if m.err != nil {
		textBlock = textStyleCentered.Render("Failed to transfer: ", m.err.Error())
	} else {
		textBlock = textStyleCentered.Render(fmt.Sprintf("TxId: %s\nAmount: %f\nFee: %f\nAddress: %s",
			m.tx.TxHash, float64(m.fiat)/100, 0.0002, m.address))
	}
	timerBlock := textStyleCentered.Render("Returning in", m.timer.View())

	doneButton := doneButtonStyle.Render("Done")

	return lipgloss.JoinVertical(lipgloss.Center, textBlock, timerBlock, zone.Mark("done", doneButton))
}
