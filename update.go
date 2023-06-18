package main

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
	"github.com/rs/zerolog/log"
	"gitlab.com/openkiosk/proto"
)

func (m model) IdleNext() (tea.Model, tea.Cmd) {
	m.state += 1
	m.timer = timer.NewWithInterval(timeout, time.Second)
	return m, tea.Batch(textinput.Blink, m.timer.Init(), waitForActivity())
}

func parseAddress(addr string) string {
	_, addressWithParams, found0 := strings.Cut(addr, ":")
	if !found0 {
		return addr
	}
	address, _, found1 := strings.Cut(addressWithParams, "?")
	if !found1 {
		return addressWithParams
	}
	return address
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
			addr := parseAddress(string(decoded))
			m.textinput.SetValue(addr)
			return m.IdleNext()
		}
		return m, waitForActivity()
	}
	return m, nil
}

func (m model) AddressInNext(s string) (tea.Model, tea.Cmd) {
	m.textinput.Reset()
	m.address = s
	cmd(m.broker, "codescannerd", "stop")
	cmd(m.broker, "pulseacceptord", "start")
	m.timer = timer.NewWithInterval(timeout, time.Second)
	pricePause <- true
	m.state += 1
	log.Info().Msg(m.address)
	return m, tea.Batch(m.spinner.Tick, m.timer.Init(), waitForActivity())
}

func addressValidator(s string) error {
	if len(s) != 95 {
		return fmt.Errorf("Invalid address length")
	}
	if cfg.Mode == "mainnet" && !(s[0] == '8' || s[0] == '4') {
		return fmt.Errorf("Invalid mainnet address")
	}
	if cfg.Mode == "stagenet" && !(s[0] == '7' || s[0] == '5') {
		return fmt.Errorf("Invalid stagenet address")
	}
	return nil
}

func (m model) AddressInUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter {
			s := m.textinput.Value()
			err := addressValidator(s)
			if err != nil {
				m.err = err
				return m, nil
			}
			return m.AddressInNext(s)
		}
	case tea.MouseMsg:
		if msg.Type != tea.MouseLeft {
			return m, nil
		}
		if zone.Get("next").InBounds(msg) {
			s := m.textinput.Value()
			err := addressValidator(s)
			if err != nil {
				m.err = err
				return m, nil
			}
			return m.AddressInNext(s)
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
			addr := parseAddress(string(decoded))
			m.textinput.SetValue(addr)
		}
		return m, waitForActivity()
	}
	var tiCmd tea.Cmd
	m.textinput, tiCmd = m.textinput.Update(msg)
	return m, tiCmd
}

func (m model) MoneyInNext() (tea.Model, tea.Cmd) {
	if m.fiat == 0 {
		return m, nil
	}
	m.state += 1
	cmd(m.broker, "pulseacceptord", "stop")
	m.timer = timer.NewWithInterval(1*time.Minute, time.Second)
	m.tx, m.err = mpayTransfer(m.xmr, m.address)
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
			m.fiat += uint64(data.Amount)
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
	m.fiat = 0
	m.xmr = 0
	m.fee = 0
	m.textinput.Reset()
	m.err = nil
	pricePause <- false
	cmd(m.broker, "pulseacceptord", "stop")
	cmd(m.broker, "codescannerd", "start")
	return m, nil
}
