package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"gitlab.com/openkiosk/proto"
	"github.com/eclipse/paho.golang/autopaho"
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

/* This where we will listen for atm backend and dispatch events
func listenForActivity(sub chan proto.Event) tea.Cmd {
	return func() tea.Msg {
		for {
			time.Sleep(time.Millisecond * time.Duration(rand.Int63n(900)+100))
			sub <- struct{}{}
		}
	}
}
*/

// A command that waits for the activity on a channel.
func waitForActivity(sub chan proto.Event) tea.Cmd {
	return func() tea.Msg {
		return <-sub
	}
}
func main() {
	// Log to a file. Useful in debugging since you can't really log to stdout.
	// Not required.
	logfilePath := os.Getenv("BUBBLETEA_LOG")
	if logfilePath != "" {
		if _, err := tea.LogToFile(logfilePath, "simple"); err != nil {
			log.Fatal(err)
		}
	}

	rand.Seed(time.Now().UTC().UnixNano())

	// Initialize our program
	p := tea.NewProgram(InitialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

// Init optionally returns an initial command we should run. In this case we
// want to start the timer.
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
	return tea.Batch(tea.EnterAltScreen,
		//listenForActivity(m.sub), // generate activity (this will come from websocket)
		waitForActivity(sub))   // wait for activity
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
				cmd(m.broker, "coinacceptord", "start")
			}
			if m.state == MoneyIn {
				cmd(m.broker, "coinacceptord", "start")
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
		if msg.Event == "moneyin" {
			m.euro += msg.Data.(proto.EventMoneyinData).Amount // record external activity
		}
		return m, waitForActivity(sub) // wait for next event
	}
	// TODO: handle signals from websocket goroutine here.
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
	return fmt.Sprintf("Euros received: %f\n\nPress enter to proceed.", m.euro)
}

func TxInfoView(m model) string {
	return fmt.Sprintf("No TxId yet but your address: %s, amount: %f EUR", m.address, m.euro)
}
