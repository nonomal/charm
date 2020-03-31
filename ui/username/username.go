package username

import (
	"errors"

	"github.com/charmbracelet/charm"
	"github.com/charmbracelet/tea"
	"github.com/charmbracelet/teaparty/input"
	te "github.com/muesli/termenv"
)

const (
	prompt = "> "
)

var (
	color         = te.ColorProfile().Color
	magenta       = "#EE6FF8"
	focusedPrompt = te.String(prompt).Foreground(color(magenta)).String()
)

type state int

const (
	nameNotChosen state = iota
	nameTaken
	nameInvalid
	nameSet
	unknownError
)

type index int

const (
	textInput index = iota
	okButton
	cancelButton
)

// MSG

type NameSetMsg struct{}

type ErrorMsg error

type ExitMsg struct{}

// MODEL

type Model struct {
	cc      *charm.Client
	state   state
	newName string
	input   input.Model
	index   index
	err     error
}

// INIT

func NewModel(cc *charm.Client) Model {
	inputModel := input.DefaultModel()
	inputModel.CursorColor = magenta
	inputModel.Placeholder = "divagurl2000"
	inputModel.Focus()
	inputModel.Prompt = focusedPrompt

	return Model{
		cc:      cc,
		state:   nameNotChosen,
		newName: "",
		input:   inputModel,
		index:   textInput,
		err:     nil,
	}
}

// UPDATE

func Update(msg tea.Msg, m Model) (Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch key := msg.Type; key {

		case tea.KeyTab:
			fallthrough
		case tea.KeyShiftTab:
			if key == tea.KeyTab {
				m.index++
				if m.index > cancelButton {
					m.index = textInput
				}
			} else {
				m.index--
				if m.index < textInput {
					m.index = cancelButton
				}
			}

			if m.index == textInput && !m.input.Focused() {
				m.input.Focus()
				m.input.Prompt = focusedPrompt

			} else if m.index != textInput && m.input.Focused() {
				m.input.Blur()
				m.input.Prompt = prompt
			}

			return m, nil

		default:
			if m.index == textInput {
				var cmd tea.Cmd
				m.input, cmd = input.Update(msg, m.input)
				return m, cmd
			}
			return m, nil
		}

	case ErrorMsg:
		switch msg {
		case charm.ErrNameTaken:
			m.state = nameTaken
			return m, nil
		default:
			m.state = unknownError
			err, ok := msg.(error)
			if !ok {
				m.err = errors.New("very, very unknown error")
			}
			m.err = err
			return m, nil
		}

	case NameSetMsg:
		m.state = nameSet
		return m, nil

	default:
		m.input, _ = input.Update(msg, m.input)
		return m, nil
	}
}

// VIEWS

func View(m Model) string {
	switch m.state {
	case nameNotChosen:
		return setNameView(m)
	default:
		return ""
	}
}

func setNameView(m Model) string {
	s := "Enter a new username\n\n"
	s += input.View(m.input) + "\n\n"
	s += buttonView("  OK  ", m.index == 1) + " " + buttonView("Cancel", m.index == 2)
	return s
}

func buttonView(label string, active bool) string {
	s := "  " + label + "  "
	c := "238"
	if active {
		c = magenta
	}
	return te.String(s).Background(color(c)).String()
}

func nameSetView(m Model) string {
	return "OK! Your new username is " + m.newName
}

// SUBSCRIPTIONS

// Blink wraps input's Blink subscription
func Blink(model tea.Model) tea.Sub {
	m, ok := model.(Model)
	if !ok {
		// TODO: handle this error properly
		return nil
	}
	return func(_ tea.Model) tea.Msg {
		return input.Blink(m.input)
	}
}

// COMMANDS

func setName(model tea.Model) tea.Msg {
	m, ok := model.(Model)
	if !ok {
		return tea.ModelAssertionErr
	}

	_, err := m.cc.SetName(m.newName)
	if err != nil {
		return ErrorMsg(err)
	}
	return NameSetMsg{}
}