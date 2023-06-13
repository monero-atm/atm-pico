package main

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

var timeout = 30 * time.Second

type model struct {
	// Timer to automatically go back to idle state if the user left it alone
	timer     timer.Model
	showTimer bool

	broker    *autopaho.ConnectionManager
	state     State
	address   string
	euro      int64
	xmr       int64
	height    int
	width     int
	textinput textinput.Model
	spinner   spinner.Model
	textstyle lipgloss.Style
	timerstyle lipgloss.Style
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
		timer:     timer.NewWithInterval(timeout, time.Second),
		broker:    connectToBroker(),
		state:     Idle,
		textinput: ti,
		textstyle: lipgloss.NewStyle().Align(lipgloss.Center).Padding(2),
		timerstyle: lipgloss.NewStyle().Align(lipgloss.Bottom).Padding(2),
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
		m.width = msg.Width
		m.height = msg.Height
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
			return m, m.timer.Init()
		}
	}
	return m, nil
}

func (m model) AddressInUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter {
			cmd(m.broker, "codescannerd", "stop")
			m.address = m.textinput.Value()
			m.textinput.Reset()
			cmd(m.broker, "pulseacceptord", "start")
			m.timer = timer.NewWithInterval(timeout, time.Second)
			NextState(&m)
			return m, tea.Batch(m.spinner.Tick, m.timer.Init())
		}
	case timer.TickMsg:
		var timerCmd tea.Cmd
		m.timer, timerCmd = m.timer.Update(msg)
		return m, timerCmd
	case timer.TimeoutMsg:
		m.state = Idle
		m.timer = timer.NewWithInterval(timeout, time.Second)
		return m, nil
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
	case timer.TickMsg:
		var timerCmd tea.Cmd
		m.timer, timerCmd = m.timer.Update(msg)
		return m, timerCmd
	case timer.TimeoutMsg:
		m.state = Idle
		return m, nil
	/*case timer.TickMsg:
		var timerCmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, timerCmd*/
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
		return m, waitForActivity(sub)
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
		}/*
	case timer.TickMsg:
		var timerCmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, timerCmd*/
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
	textBlock := m.textstyle.Render(
                "Enter the receiving address using the keyboard or scan QR code:\n\n",
                m.textinput.View())
	//tBlock := m.timerstyle.Render(m.timer.View())
	h := lipgloss.Height(textBlock)
	tBlock := lipgloss.Place(m.width, m.height-h, lipgloss.Right, lipgloss.Bottom, m.timer.View())
	return lipgloss.JoinVertical(lipgloss.Right, textBlock, tBlock)
}

func MoneyInView(m model) string {
	textBlock := m.textstyle.Render(m.spinner.View(),
		fmt.Sprintf("Received: %.2f EUR", float64(m.euro)),
		"\n\n", "Press enter to proceed.")

	h := lipgloss.Height(textBlock)
	tBlock := lipgloss.Place(m.width, m.height-h, lipgloss.Right, lipgloss.Bottom, m.timer.View())
	return lipgloss.JoinVertical(lipgloss.Right, textBlock, tBlock)
}

func TxInfoView(m model) string {
	return m.textstyle.Render(fmt.Sprintf("No TxId yet but your address: %s, amount: %.2f EUR", m.address, float64(m.euro)))
}
