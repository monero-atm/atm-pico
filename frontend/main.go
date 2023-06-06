package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type State int

const (
	Idle State = iota
	AddressIn
	MoneyIn
	TxInfo
)

type model struct {
	// sub is the ATM backend websocket event channel. For now it's an
	// anonymous struct since we don't have a protocol spec/type def.
	sub       chan struct{}
	state     State
	address   string
	euro      float64
	xmr       float64
	textinput textinput.Model
}

// A message used to indicate that activity has occurred. When we have
// a protocol spec it will contain related data
type responseMsg struct{}

// This where we will listen for atm backend and dispatch events
func listenForActivity(sub chan struct{}) tea.Cmd {
	return func() tea.Msg {
		for {
			time.Sleep(time.Millisecond * time.Duration(rand.Int63n(900)+100))
			sub <- struct{}{}
		}
	}
}

// A command that waits for the activity on a channel.
func waitForActivity(sub chan struct{}) tea.Cmd {
	return func() tea.Msg {
		return responseMsg(<-sub)
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

	return model{
		sub:       make(chan struct{}),
		state:     Idle,
		textinput: ti,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen,
		listenForActivity(m.sub), // generate activity (this will come from websocket)
		waitForActivity(m.sub))   // wait for activity
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			// proceed to next state
			if m.state == AddressIn {
				m.address = m.textinput.Value()
				m.textinput.Reset()
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
	case responseMsg:
		if m.state == MoneyIn {
			m.euro++ // record external activity
		}
		return m, waitForActivity(m.sub) // wait for next event
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
