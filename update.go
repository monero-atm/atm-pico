package main

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rs/zerolog/log"
	"gitlab.com/openkiosk/proto"
)

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
		return m, waitForActivity()
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
		return m, waitForActivity()
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
		} /*
			case timer.TickMsg:
				var timerCmd tea.Cmd
				m.timer, cmd = m.timer.Update(msg)
				return m, timerCmd*/
	}
	return m, nil
}
