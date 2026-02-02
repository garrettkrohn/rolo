package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type mode int

const (
	normalMode mode = iota
	moveMode
)

type model struct {
	sessions   []string
	cursor     int
	mode       mode
	onSave     func([]string) error
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "m":
			// Toggle move mode
			if m.mode == normalMode {
				m.mode = moveMode
			} else {
				m.mode = normalMode
			}

		case "j":
			if m.mode == normalMode {
				// Move cursor down
				if m.cursor < len(m.sessions)-1 {
					m.cursor++
				}
			} else {
				// Move item down
				if m.cursor < len(m.sessions)-1 {
					m.sessions[m.cursor], m.sessions[m.cursor+1] =
						m.sessions[m.cursor+1], m.sessions[m.cursor]
					m.cursor++
				}
			}

		case "k":
			if m.mode == normalMode {
				// Move cursor up
				if m.cursor > 0 {
					m.cursor--
				}
			} else {
				// Move item up
				if m.cursor > 0 {
					m.sessions[m.cursor], m.sessions[m.cursor-1] =
						m.sessions[m.cursor-1], m.sessions[m.cursor]
					m.cursor--
				}
			}

		case "enter":
			// Save and quit
			if m.onSave != nil {
				if err := m.onSave(m.sessions); err != nil {
					// Could add error handling here
					return m, tea.Quit
				}
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	s := "Reorder tmux sessions\n"
	if m.mode == moveMode {
		s += "[MOVE MODE] - j/k to move item, m to exit move mode\n\n"
	} else {
		s += "[NORMAL] - j/k to navigate, m to enter move mode, enter to save\n\n"
	}

	for i, session := range m.sessions {
		cursor := "  "
		if m.cursor == i {
			if m.mode == moveMode {
				cursor = "▶ "
			} else {
				cursor = "> "
			}
		}
		s += fmt.Sprintf("%s%s\n", cursor, session)
	}
	return s
}

// Run starts the interactive TUI for reordering sessions
func Run(sessions []string, onSave func([]string) error) error {
	m := model{
		sessions: sessions,
		cursor:   0,
		mode:     normalMode,
		onSave:   onSave,
	}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	return nil
}
