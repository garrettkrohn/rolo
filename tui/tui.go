package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"rolo/storage"
	"rolo/tmux"
)

// Catppuccin Mocha color palette
var (
	catppuccinRosewater = lipgloss.Color("#f5e0dc")
	catppuccinFlamingo  = lipgloss.Color("#f2cdcd")
	catppuccinPink      = lipgloss.Color("#f5c2e7")
	catppuccinMauve     = lipgloss.Color("#cba6f7")
	catppuccinRed       = lipgloss.Color("#f38ba8")
	catppuccinMaroon    = lipgloss.Color("#eba0ac")
	catppuccinPeach     = lipgloss.Color("#fab387")
	catppuccinYellow    = lipgloss.Color("#f9e2af")
	catppuccinGreen     = lipgloss.Color("#a6e3a1")
	catppuccinTeal      = lipgloss.Color("#94e2d5")
	catppuccinSky       = lipgloss.Color("#89dceb")
	catppuccinSapphire  = lipgloss.Color("#74c7ec")
	catppuccinBlue      = lipgloss.Color("#89b4fa")
	catppuccinLavender  = lipgloss.Color("#b4befe")
	catppuccinText      = lipgloss.Color("#cdd6f4")
	catppuccinSubtext1  = lipgloss.Color("#bac2de")
	catppuccinSubtext0  = lipgloss.Color("#a6adc8")
	catppuccinOverlay2  = lipgloss.Color("#9399b2")
	catppuccinOverlay1  = lipgloss.Color("#7f849c")
	catppuccinOverlay0  = lipgloss.Color("#6c7086")
	catppuccinSurface2  = lipgloss.Color("#585b70")
	catppuccinSurface1  = lipgloss.Color("#45475a")
	catppuccinSurface0  = lipgloss.Color("#313244")
	catppuccinBase      = lipgloss.Color("#1e1e2e")
	catppuccinMantle    = lipgloss.Color("#181825")
	catppuccinCrust     = lipgloss.Color("#11111b")
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
	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(catppuccinMauve).
		Background(catppuccinSurface0).
		Padding(0, 2).
		MarginBottom(1)
	
	modeNormalStyle := lipgloss.NewStyle().
		Foreground(catppuccinGreen).
		Bold(true)
	
	modeMoveStyle := lipgloss.NewStyle().
		Foreground(catppuccinPeach).
		Bold(true)
	
	helpStyle := lipgloss.NewStyle().
		Foreground(catppuccinSubtext0).
		Italic(true)
	
	keybindStyle := lipgloss.NewStyle().
		Foreground(catppuccinBlue).
		Bold(true)
	
	cursorNormalStyle := lipgloss.NewStyle().
		Foreground(catppuccinPink).
		Bold(true)
	
	cursorMoveStyle := lipgloss.NewStyle().
		Foreground(catppuccinPeach).
		Bold(true)
	
	sessionActiveStyle := lipgloss.NewStyle().
		Foreground(catppuccinText)
	
	sessionDeletedStyle := lipgloss.NewStyle().
		Foreground(catppuccinOverlay0).
		Strikethrough(true)
	
	sessionHighlightStyle := lipgloss.NewStyle().
		Foreground(catppuccinText).
		Background(catppuccinSurface0).
		Bold(true)
	
	// Build the view
	var s string
	
	// Title
	s += titleStyle.Render("✨ Rolo - Tmux Session Manager") + "\n\n"
	
	// Mode indicator and help text
	if m.mode == moveMode {
		modeText := modeMoveStyle.Render("MOVE MODE")
		help := helpStyle.Render(
			keybindStyle.Render("j/k") + " move item  " +
			keybindStyle.Render("m") + " exit move mode",
		)
		s += modeText + " - " + help + "\n\n"
	} else {
		modeText := modeNormalStyle.Render("NORMAL")
		help := helpStyle.Render(
			keybindStyle.Render("j/k") + " navigate  " +
			keybindStyle.Render("d") + " delete  " +
			keybindStyle.Render("u") + " update  " +
			keybindStyle.Render("p") + " repopulate  " +
			keybindStyle.Render("m") + " move  " +
			keybindStyle.Render("enter") + " save",
		)
		s += modeText + " - " + help + "\n\n"
	}

	// Session list
	for i, session := range m.sessions {
		var line string
		
		// Cursor indicator
		cursor := "  "
		if m.cursor == i {
			if m.mode == moveMode {
				cursor = cursorMoveStyle.Render("▶ ")
			} else {
				cursor = cursorNormalStyle.Render("› ")
			}
		}
		
		// Session name with styling
		sessionText := session.Name
		if session.Deleted {
			sessionText = sessionDeletedStyle.Render(session.Name)
		} else if m.cursor == i {
			sessionText = sessionHighlightStyle.Render(session.Name)
		} else {
			sessionText = sessionActiveStyle.Render(session.Name)
		}
		
		line = cursor + sessionText
		s += line + "\n"
	}
	
	// Footer
	s += "\n" + helpStyle.Render("Press ") + keybindStyle.Render("q") + helpStyle.Render(" or ") + keybindStyle.Render("ctrl+c") + helpStyle.Render(" to quit without saving")
	
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
