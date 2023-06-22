package main

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/eclipse/paho.golang/autopaho"
	zone "github.com/lrstanley/bubblezone"
	"github.com/rs/zerolog/log"
	mpay "gitlab.com/moneropay/moneropay/v2/pkg/model"
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
	timer     timer.Model
	broker    *autopaho.ConnectionManager
	state     State
	address   string
	fiat      uint64
	xmr       uint64
	fee       float64
	xmrPrice  float64
	err       error
	height    int
	width     int
	tx        *mpay.TransferPostResponse
	textinput textinput.Model
	spinner   spinner.Model
}

var sub chan proto.Event

var priceUpdate chan priceEvent
var pricePause chan bool

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
	pricePause = make(chan bool)
	go pricePoll()
	zone.NewGlobal()
	p := tea.NewProgram(InitialModel(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		log.Fatal().Err(err)
	}
}

func InitialModel() model {
	ti := textinput.New()
	ti.Placeholder = "..."
	ti.CharLimit = 95
	ti.Width = 95
	ti.Focus()

	sub = make(chan proto.Event)

	xp, err := getXmrPrice()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get XMR price")
		xp = cfg.FallbackPrice
	}

	m := model{
		timer:     timer.NewWithInterval(cfg.StateTimeout, time.Second),
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
	cmd(m.broker, "codescannerd", "start")
	return tea.Batch(tea.EnterAltScreen,
		waitForActivity(),
		waitForPriceUpdate())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m model) View() string {
	switch m.state {
	case AddressIn:
		return zone.Scan(AddressInView(m))
	case MoneyIn:
		return zone.Scan(MoneyInView(m))
	case TxInfo:
		return zone.Scan(TxInfoView(m))
	}
	return IdleView(m)
}
