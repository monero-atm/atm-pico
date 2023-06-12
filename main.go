package main

import (
	"encoding/base64"
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/rs/zerolog/log"
	"gitlab.com/openkiosk/proto"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/spinner"
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
	height, width int
	textinput textinput.Model
	spinner spinner.Model
	textstyle lipgloss.Style
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

	m := model{
                broker:    connectToBroker(),
                state:     Idle,
                textinput: ti,
                textstyle: lipgloss.NewStyle().Align(lipgloss.Center).Padding(2),
        }

	spinnerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("69"))

	m.spinner = spinner.New()
	m.spinner.Style = spinnerStyle
	m.spinner.Spinner = spinner.Pulse
	
	return m
}

func (m model) Init() tea.Cmd {
	m.state = Idle
	return tea.Batch(tea.EnterAltScreen,
		waitForActivity(sub)) // wait for activity
}

func NextState(m *model) {
	m.state += 1
	if m.state > 3 {
		// Reset to Idle
		m.state = Idle
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// These messages are handled always regardless of the state
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.textstyle.Width(msg.Width)
		m.textstyle.Height(msg.Height)
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			brokerDisconnect(m.broker)
			return m, tea.Quit
		}
	}

	switch m.state {
		case Idle:
			return m.IdleUpdate(msg)
		case AddressIn:
			return m.AddressInUpdate(msg)
		case MoneyIn:
			return m.MoneyInUpdate(msg)
		case TxInfo:
			return m.TxInfoUpdate(msg)
	}

	return m, nil
}

func (m model) IdleUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Info().Msg("Hello from idle")
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter {
			cmd(m.broker, "codescannerd", "start")
			NextState(&m)
			//log.Info().Msg(fmt.Sprintf("state=%v", m.state))
			return m, nil
		}
	}
	return m, nil
}

func (m model) AddressInUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	//log.Info().Msg("Hello from addressin")
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter {
			cmd(m.broker, "codescannerd", "stop")
			m.address = m.textinput.Value()
			m.textinput.Reset()
			cmd(m.broker, "pulseacceptord", "start")
			NextState(&m)
			return m, m.spinner.Tick
		}
	case proto.Event:
		log.Info().Str("type", msg.Event).Msg("Got event!")
		log.Info().Msg("case proto.Event")
	
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
		return m, waitForActivity(sub)
	}
	var tiCmd tea.Cmd
	m.textinput, tiCmd = m.textinput.Update(msg)
	return m, tiCmd
}

func (m model) MoneyInUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter {
			NextState(&m)
			cmd(m.broker, "pulseacceptord", "stop")
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
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m model) TxInfoUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter {
			NextState(&m)
		}
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
	return m.textstyle.Render("Displaying cool ads and animations. Press any key to start buying Monero.")
	//return fmt.Sprintf("Displaying cool ads and animations. Press any key to start buying Monero.")
}

func AddressInView(m model) string {
	return m.textstyle.Render(
	"Enter the receiving address using the keyboard or scan QR code:\n\n",
		m.textinput.View())
}

func MoneyInView(m model) string {
	return m.textstyle.Render(m.spinner.View(),fmt.Sprintf("Received: %.2f EUR", float64(m.euro)), "\n\n", "Press enter to proceed.")
}

func TxInfoView(m model) string {
	return m.textstyle.Render(fmt.Sprintf("No TxId yet but your address: %s, amount: %.2f EUR", m.address, float64(m.euro)))
}
