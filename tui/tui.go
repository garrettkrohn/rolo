package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"rolo/storage"
	"rolo/tmux"
)

type mode int

const (
	normalMode mode = iota
	moveMode
)

type model struct {
	sessions   []storage.SessionData
	cursor     int
	mode       mode
	onSave     func([]storage.SessionData) error
	wrapAround bool
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

		case "d":
			// Toggle deleted state for current session
			if m.cursor < len(m.sessions) {
				m.sessions[m.cursor].Deleted = !m.sessions[m.cursor].Deleted
			}

		case "p":
			// Repopulate from active tmux sessions
			sessions, err := tmux.GetActiveSessions()
			if err != nil {
				// If we can't get sessions, just keep current state
				return m, nil
			}
			
			// Convert to SessionData format (all non-deleted by default)
			sessionData := make([]storage.SessionData, len(sessions))
			for i, name := range sessions {
				sessionData[i] = storage.SessionData{Name: name, Deleted: false}
			}
			
			// Replace current sessions and reset cursor
			m.sessions = sessionData
			m.cursor = 0
			if m.cursor >= len(m.sessions) && len(m.sessions) > 0 {
				m.cursor = len(m.sessions) - 1
			}

		case "u":
			// Update list by adding new tmux sessions and removing closed ones
			sessions, err := tmux.GetActiveSessions()
			if err != nil {
				// If we can't get sessions, just keep current state
				return m, nil
			}
			
			// Create a map of active session names for quick lookup
			activeNames := make(map[string]bool)
			for _, name := range sessions {
				activeNames[name] = true
			}
			
			// Filter out sessions that are no longer active
			filteredSessions := make([]storage.SessionData, 0, len(m.sessions))
			for _, session := range m.sessions {
				if activeNames[session.Name] {
					filteredSessions = append(filteredSessions, session)
					delete(activeNames, session.Name) // Remove from map so we know it's been seen
				}
			}
			
			// Add any new sessions that weren't in the list
			for name := range activeNames {
				filteredSessions = append(filteredSessions, storage.SessionData{
					Name:    name,
					Deleted: false,
				})
			}
			
			// Update sessions and adjust cursor if needed
			m.sessions = filteredSessions
			if m.cursor >= len(m.sessions) && len(m.sessions) > 0 {
				m.cursor = len(m.sessions) - 1
			}

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
				} else if m.wrapAround && len(m.sessions) > 0 {
					m.cursor = 0
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
				} else if m.wrapAround && len(m.sessions) > 0 {
					m.cursor = len(m.sessions) - 1
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
	strikethroughStyle := lipgloss.NewStyle().Strikethrough(true)
	
	s := "Reorder tmux sessions\n"
	if m.mode == moveMode {
		s += "[MOVE MODE] - j/k to move item, m to exit move mode\n\n"
	} else {
		s += "[NORMAL] - j/k to navigate, d to delete, u to update, p to repopulate, m to enter move mode, enter to save\n\n"
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
		
		sessionText := session.Name
		if session.Deleted {
			sessionText = strikethroughStyle.Render(session.Name)
		}
		
		s += fmt.Sprintf("%s%s\n", cursor, sessionText)
	}
	return s
}

// Run starts the interactive TUI for reordering sessions
func Run(sessions []storage.SessionData, onSave func([]storage.SessionData) error) error {
	// Load config to get wrap around setting
	config, err := storage.LoadConfig()
	if err != nil {
		// If config fails to load, use default (false)
		config = &storage.Config{WrapAround: false}
	}

	m := model{
		sessions:   sessions,
		cursor:     0,
		mode:       normalMode,
		onSave:     onSave,
		wrapAround: config.WrapAround,
	}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	return nil
}
