package main

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/timer"
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
	xmrPrice  float64
	height    int
	width     int
	textinput textinput.Model
	spinner   spinner.Model
}

// MQTT events
var sub chan proto.Event

var priceUpdate chan priceEvent

func waitForActivity() tea.Cmd {
	return func() tea.Msg {
		return <-sub
	}
}

func waitForPriceUpdate() tea.Cmd {
	return func() tea.Msg {
		return <-priceUpdate
	}
}

func main() {
	cfg = loadConfig()
	priceUpdate = make(chan priceEvent) 
	go pricePoll()
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

	xp, err := getXmrPrice()
	if err != nil {
		log.Fatal().Err(err).Msg("")
		xp = 123.456
	}

	m := model{
		timer:     timer.NewWithInterval(timeout, time.Second),
		broker:    connectToBroker(),
		state:     Idle,
		textinput: ti,
		xmrPrice:  xp,
	}

	m.spinner = spinner.New()
	m.spinner.Style = spinnerStyle
	m.spinner.Spinner = spinner.Pulse

	return m
}

func (m model) Init() tea.Cmd {
	m.state = Idle
	return tea.Batch(tea.EnterAltScreen,
		waitForActivity(),
		waitForPriceUpdate())
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
		textStyleCentered.Width(msg.Width)
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			brokerDisconnect(m.broker)
			return m, tea.Quit
		}
	case priceEvent:
		log.Info().Float64("rate", float64(msg)).Msg("Got price update!")
		m.xmrPrice = float64(msg)
		return m, waitForPriceUpdate()
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
