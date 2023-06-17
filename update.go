package main

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
	"github.com/rs/zerolog/log"
	"gitlab.com/openkiosk/proto"
)

func (m model) IdleNext() (tea.Model, tea.Cmd) {
	m.state += 1
	m.timer = timer.NewWithInterval(timeout, time.Second)
	return m, m.timer.Init()
}

func (m model) IdleUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter {
			return m.IdleNext()
		}
	case tea.MouseMsg:
		if msg.Type != tea.MouseLeft {
			return m, nil
		}
		return m.IdleNext()
	case proto.Event:
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
			return m.IdleNext()
		}
		return m, waitForActivity()

	}
	return m, nil
}

func (m model) AddressInNext() (tea.Model, tea.Cmd) {
	cmd(m.broker, "codescannerd", "stop")
	m.address = m.textinput.Value()
	m.textinput.Reset()
	cmd(m.broker, "pulseacceptord", "start")
	m.timer = timer.NewWithInterval(timeout, time.Second)
	pricePause <- true
	m.state += 1
	return m, tea.Batch(m.spinner.Tick, m.timer.Init())
}

func (m model) AddressInUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter {
			return m.AddressInNext()
		}
	case tea.MouseMsg:
		if msg.Type != tea.MouseLeft {
			return m, nil
		}
		if zone.Get("next").InBounds(msg) {
			return m.AddressInNext()
		} else if zone.Get("back").InBounds(msg) {
			return m.BackToIdle()
		}
	case timer.TickMsg:
		var timerCmd tea.Cmd
		m.timer, timerCmd = m.timer.Update(msg)
		return m, timerCmd
	case timer.TimeoutMsg:
		return m.BackToIdle()
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

func (m model) MoneyInNext() (tea.Model, tea.Cmd) {
	m.state += 1
	cmd(m.broker, "pulseacceptord", "stop")
	m.timer = timer.NewWithInterval(1 * time.Minute, time.Second)
	return m, m.timer.Init()
}

func (m model) MoneyInUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter {
			return m.MoneyInNext()
		}
	case tea.MouseMsg:
		if msg.Type != tea.MouseLeft {
			return m, nil
		}
		if zone.Get("next").InBounds(msg) {
			return m.MoneyInNext()
		} else if zone.Get("back").InBounds(msg) {
			return m.BackToIdle()
		}
	case timer.TickMsg:
		var timerCmd tea.Cmd
		m.timer, timerCmd = m.timer.Update(msg)
		return m, timerCmd
	case timer.TimeoutMsg:
		return m.BackToIdle()
	case proto.Event:
		log.Info().Str("type", msg.Event).Msg("Got event!")
		log.Info().Msg("case proto.Event")
		if msg.Event == "moneyin" {
			data, err := proto.GetMoneyinData(msg.Data)
			if err != nil {
				log.Error().Err(err).Msg("Failed to unmarshall scan data")
			}
			m.euro += data.Amount
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
			return m.BackToIdle()
		}
	case timer.TickMsg:
		var timerCmd tea.Cmd
		m.timer, timerCmd = m.timer.Update(msg)
		return m, timerCmd
	case timer.TimeoutMsg:
		return m.BackToIdle()
	case tea.MouseMsg:
		if msg.Type != tea.MouseLeft {
			return m, nil
		}
		if zone.Get("next").InBounds(msg) {
			return m.MoneyInNext()
		} else if zone.Get("done").InBounds(msg) {
			return m.BackToIdle()
		}
	}
	return m, nil
}

func (m model) BackToIdle() (tea.Model, tea.Cmd) {
	m.state = Idle
	m.address = ""
	m.euro = 0
	m.xmr = 0
	m.fee = 0
	m.textinput.Reset()
	pricePause <- false
	return m, nil
}
