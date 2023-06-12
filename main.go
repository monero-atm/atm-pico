package main

import (
	"encoding/base64"
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/rs/zerolog/log"
	"gitlab.com/openkiosk/proto"
)

type State int

const (
	Idle State = iota
	AddressIn
	MoneyIn
	TxInfo
)

type model struct {
	broker    *autopaho.ConnectionManager
	state     State
	address   string
	euro      int64
	xmr       int64
	textinput textinput.Model
}

var sub chan proto.Event

// A command that waits for the activity on a channel.
func waitForActivity(sub chan proto.Event) tea.Cmd {
	return func() tea.Msg {
		log.Info().Msg("waitForActivity")
		return <-sub
	}
}
func main() {
	p := tea.NewProgram(InitialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal().Err(err)
	}
}

func InitialModel() model {
	ti := textinput.New()
	ti.Placeholder = "8..."
	ti.Focus()
	// TODO: add input validator function here for address

	sub = make(chan proto.Event)

	return model{
		broker:    connectToBroker(),
		state:     Idle,
		textinput: ti,
	}
}

func (m model) Init() tea.Cmd {
	m.state = Idle
	return tea.Batch(tea.EnterAltScreen,
		waitForActivity(sub)) // wait for activity
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			brokerDisconnect(m.broker)
			return m, tea.Quit
		case tea.KeyEnter:
			// proceed to next state
			if m.state == Idle {
				cmd(m.broker, "codescannerd", "start")
			}

			if m.state == AddressIn {
				cmd(m.broker, "codescannerd", "stop")
				m.address = m.textinput.Value()
				m.textinput.Reset()
				cmd(m.broker, "pulseacceptord", "start")
			}
			if m.state == MoneyIn {
				cmd(m.broker, "pulseacceptord", "stop")
			}

			m.state += 1
			//return m, tea.Quit
		}
		if m.state == AddressIn {
			var tiCmd tea.Cmd
			m.textinput, tiCmd = m.textinput.Update(msg)
			return m, tea.Batch(tiCmd)
		}
		if m.state > 3 {
			// Reset to Idle
			m.state = Idle
			return m, nil
		}
	case proto.Event:
		log.Info().Str("type", msg.Event).Msg("Got event!")
		log.Info().Msg("case proto.Event")
		if msg.Event == "moneyin" {
			data, err := proto.GetMoneyinData(msg.Data)
			if err != nil {
				log.Error().Err(err).Msg("Failed to unmarshall scan data")
			}
			m.euro += data.Amount // record external activity
		}
		if msg.Event == "codescan" {
			log.Info().Str("data", fmt.Sprintf("%v", msg)).Msg("")
			data, err := proto.GetScanData(msg.Data)
			if err != nil {
				log.Error().Err(err).Msg("Failed to unmarshall scan data")
			}
			decoded, err := base64.StdEncoding.DecodeString(data.Scan)
			if err != nil {
				panic(err)
			}
			m.textinput.SetValue(string(decoded))
		}
		return m, waitForActivity(sub) // wait for next event
	}
	return m, nil
}

// View returns a string based on data in the model. That string which will be
// rendered to the terminal.
func (m model) View() string {
	switch m.state {
	case AddressIn:
		return AddressInView(m)
	case MoneyIn:
		return MoneyInView(m)
	case TxInfo:
		return TxInfoView(m)
	}
	return IdleView(m)
}

// TODO: Pimp my idle view
func IdleView(m model) string {
	return fmt.Sprintf("Displaying cool ads and animations. Press any key to start buying Monero.")
}

func AddressInView(m model) string {
	return fmt.Sprintf("Enter the receiving address using the keyboard or scan QR code\n\n%s", m.textinput.View())
}

func MoneyInView(m model) string {
	return fmt.Sprintf("Received: %.2f EUR\n\nPress enter to proceed.", float64(m.euro))
}

func TxInfoView(m model) string {
	return fmt.Sprintf("No TxId yet but your address: %s, amount: %.2f EUR", m.address, float64(m.euro))
}
