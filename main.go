package main

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var timeout = time.Minute * 25

var percent float64 = 0.0

const (
	padding  = 2
	maxWidth = 80
)

type model struct {
	timer    timer.Model
	keymap   keymap
	help     help.Model
	quitting bool
	progress progress.Model
}

type tickMsg time.Time

type keymap struct {
	start      key.Binding
	pauseTimer key.Binding
	workTimer  key.Binding
	stop       key.Binding
	reset      key.Binding
	quit       key.Binding
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		m.timer.Stop(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case timer.TickMsg:
		var cmd tea.Cmd
		percent = ((timeout.Seconds() - m.timer.Timeout.Seconds()) / timeout.Seconds())

		progressCmd := m.progress.SetPercent(float64(percent))
		m.timer, cmd = m.timer.Update(msg)

		return m, tea.Batch(progressCmd, cmd)

	case timer.StartStopMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		m.keymap.stop.SetEnabled(m.timer.Running())
		m.keymap.start.SetEnabled(!m.timer.Running())
		return m, cmd

	case timer.TimeoutMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		m.quitting = true
		m.keymap.stop.SetEnabled(m.timer.Running())
		m.keymap.start.SetEnabled(!m.timer.Running())
		return m, cmd

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.quit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, m.keymap.reset):
			progressCmd := m.progress.SetPercent(0.0)
			m.timer = timer.New(timeout)

			m.keymap.start.SetEnabled(true)

			return m, tea.Batch(progressCmd, m.timer.Stop())
		case key.Matches(msg, m.keymap.start, m.keymap.stop):
			return m, m.timer.Toggle()
		case key.Matches(msg, m.keymap.pauseTimer):
			progressCmd := m.progress.SetPercent(0.0)
			pause := time.Minute * 5
			m.timer = timer.New(pause)
			return m, tea.Batch(progressCmd, m.timer.Start())
		case key.Matches(msg, m.keymap.workTimer):
			progressCmd := m.progress.SetPercent(0.0)
			timeout = time.Minute * 25
			m.timer = timer.New(timeout)
			return m, tea.Batch(progressCmd, m.timer.Start())
		}

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
	case tea.WindowSizeMsg:
		m.progress.Width = msg.Width - padding*2 - 4
		if m.progress.Width > maxWidth {
			m.progress.Width = maxWidth
		}
		return m, nil
	default:
		return m, nil
	}

	return m, nil
}

func (m model) helpView() string {
	return "\n" + m.help.ShortHelpView([]key.Binding{
		m.keymap.start,
		m.keymap.stop,
		m.keymap.reset,
		m.keymap.quit,
		m.keymap.pauseTimer,
		m.keymap.workTimer,
	})
}

func (m model) View() string {
	s := m.timer.View()

	if m.timer.Timedout() {
		s = "All done!"
	}

	var style = lipgloss.NewStyle().
		Bold(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		PaddingLeft(1).
		PaddingRight(1).
		PaddingTop(1).
		BorderTop(false).
		BorderBottom(true).
		BorderLeft(false).
		BorderRight(false)

	var textStyle = lipgloss.NewStyle().
		Bold(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		PaddingLeft(2).
		Width(82).
		PaddingBottom(1).
		BorderTop(true).
		Foreground(lipgloss.Color("63"))

	prog := m.progress.View() + m.helpView()

	return (textStyle.Render(s) + style.Render(prog))
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Minute*1, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func main() {
	m := model{
		timer: timer.New(timeout),
		progress: progress.New(progress.WithDefaultGradient(),
			progress.WithWidth(40),
			progress.WithoutPercentage()),
		keymap: keymap{
			start: key.NewBinding(
				key.WithKeys("s", " "),
				key.WithHelp("s", "start"),
			),
			stop: key.NewBinding(
				key.WithKeys("s", " "),
				key.WithHelp("s", "stop"),
			),
			reset: key.NewBinding(
				key.WithKeys("r"),
				key.WithHelp("r", "reset"),
			),
			quit: key.NewBinding(
				key.WithKeys("q", "ctrl+c"),
				key.WithHelp("q", "quit"),
			),
			pauseTimer: key.NewBinding(
				key.WithKeys("p"),
				key.WithHelp("p", "start break"),
			),
			workTimer: key.NewBinding(
				key.WithKeys("w"),
				key.WithHelp("w", "start work"),
			),
		},
		help: help.New(),
	}

	m.keymap.stop.SetEnabled(false)

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Uh oh, we encountered an error:", err)
		os.Exit(1)
	}
}
