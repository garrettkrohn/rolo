package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
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
	inputMode
	confirmMode
)

type inputAction int

const (
	createGroupAction inputAction = iota
	renameGroupAction
)

type model struct {
	sessions      []storage.SessionData
	cursor        int
	mode          mode
	onSave        func([]storage.SessionData) error
	wrapAround    bool
	groupsData    *storage.GroupsData
	textInput     textinput.Model
	inputAction   inputAction
	confirmPrompt string
	showDeleted   bool
	config        *storage.Config
	showKeymap    bool
}

func (m model) Init() tea.Cmd {
	return nil
}

// isSessionVisible returns true if a session should be visible in the list
func (m model) isSessionVisible(index int) bool {
	if index < 0 || index >= len(m.sessions) {
		return false
	}
	// Session is visible if it's not deleted, or if we're showing deleted sessions
	return !m.sessions[index].Deleted || m.showDeleted
}

// findNextVisibleSession finds the next visible session from the current position
// Returns -1 if no visible session is found
func (m model) findNextVisibleSession(start int) int {
	for i := start + 1; i < len(m.sessions); i++ {
		if m.isSessionVisible(i) {
			return i
		}
	}
	// If wrapAround is enabled, search from the beginning
	if m.wrapAround {
		for i := 0; i < start; i++ {
			if m.isSessionVisible(i) {
				return i
			}
		}
	}
	return -1
}

// findPrevVisibleSession finds the previous visible session from the current position
// Returns -1 if no visible session is found
func (m model) findPrevVisibleSession(start int) int {
	for i := start - 1; i >= 0; i-- {
		if m.isSessionVisible(i) {
			return i
		}
	}
	// If wrapAround is enabled, search from the end
	if m.wrapAround {
		for i := len(m.sessions) - 1; i > start; i-- {
			if m.isSessionVisible(i) {
				return i
			}
		}
	}
	return -1
}

// ensureCursorOnVisibleSession moves cursor to nearest visible session if current position is hidden
func (m *model) ensureCursorOnVisibleSession() {
	if len(m.sessions) == 0 {
		m.cursor = 0
		return
	}

	// If current position is visible, we're good
	if m.isSessionVisible(m.cursor) {
		return
	}

	// Try to find next visible session
	next := m.findNextVisibleSession(m.cursor)
	if next != -1 {
		m.cursor = next
		return
	}

	// Try to find previous visible session
	prev := m.findPrevVisibleSession(m.cursor)
	if prev != -1 {
		m.cursor = prev
		return
	}

	// No visible sessions at all, reset cursor to 0
	m.cursor = 0
}

// getAllGroupNames returns all group names with indicator for current group
func getAllGroupNames(groupsData *storage.GroupsData) string {
	if groupsData == nil || len(groupsData.Groups) == 0 {
		return ""
	}

	var parts []string
	for i, group := range groupsData.Groups {
		if i == groupsData.CurrentGroup {
			parts = append(parts, "👉 "+group.Name)
		} else {
			parts = append(parts, group.Name)
		}
	}

	return strings.Join(parts, " | ")
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle confirm mode
		if m.mode == confirmMode {
			switch msg.String() {
			case "y", "Y":
				// Confirm delete
				if len(m.groupsData.Groups) > 1 {
					storage.DeleteGroup(m.groupsData, m.groupsData.CurrentGroup)
					// Update sessions to show new current group
					m.sessions = m.groupsData.Groups[m.groupsData.CurrentGroup].Sessions
					m.cursor = 0
				}
				m.mode = normalMode
				return m, nil
			case "n", "N", "esc":
				// Cancel
				m.mode = normalMode
				return m, nil
			}
			return m, nil
		}

		// Handle input mode separately
		if m.mode == inputMode {
			switch msg.String() {
			case "esc":
				// Cancel input
				m.mode = normalMode
				return m, nil
			case "enter":
				// Submit input
				inputValue := strings.TrimSpace(m.textInput.Value())

				if m.inputAction == createGroupAction {
					if err := storage.CreateGroup(m.groupsData, inputValue); err == nil {
						// Successfully created group
						m.mode = normalMode
					}
					// If error, stay in input mode so user can fix it
				} else if m.inputAction == renameGroupAction {
					if err := storage.RenameGroup(m.groupsData, m.groupsData.CurrentGroup, inputValue); err == nil {
						m.mode = normalMode
					}
				}
				return m, nil
			default:
				// Update text input
				var cmd tea.Cmd
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
		}

		// Normal mode and move mode handling
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.mode == moveMode {
				m.mode = normalMode
			}
			return m, nil

		case "o":
			// Attach to the session under cursor
			if m.cursor < len(m.sessions) {
				sessionName := m.sessions[m.cursor].Name
				if err := tmux.AttachToSession(sessionName); err != nil {
					// If attach fails, just continue
					return m, nil
				}
				// Save before attaching
				if m.onSave != nil {
					m.onSave(m.sessions)
				}
				return m, tea.Quit
			}

		case "?":
			// Toggle keymap visibility
			m.showKeymap = !m.showKeymap

		case "t":
			// Toggle showing deleted sessions
			m.showDeleted = !m.showDeleted
			// Save the preference to config
			if m.config != nil {
				m.config.ShowDeleted = m.showDeleted
				storage.SaveConfig(m.config)
			}
			// Ensure cursor is on a visible session
			m.ensureCursorOnVisibleSession()

		case "d":
			// Toggle deleted state for current session
			if m.cursor < len(m.sessions) {
				m.sessions[m.cursor].Deleted = !m.sessions[m.cursor].Deleted
			}

		case "D":
			// Kill the tmux session under cursor
			if m.cursor < len(m.sessions) {
				sessionName := m.sessions[m.cursor].Name
				if err := tmux.KillSession(sessionName); err != nil {
					// If kill fails, just mark as deleted
					m.sessions[m.cursor].Deleted = true
				} else {
					// Remove the session from the list
					m.sessions = append(m.sessions[:m.cursor], m.sessions[m.cursor+1:]...)
					// Adjust cursor if needed
					if m.cursor >= len(m.sessions) && len(m.sessions) > 0 {
						m.cursor = len(m.sessions) - 1
					}
				}
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
				// Move cursor down to next visible session
				next := m.findNextVisibleSession(m.cursor)
				if next != -1 {
					m.cursor = next
				}
			} else {
				// Move item down - swap with next visible session
				next := m.findNextVisibleSession(m.cursor)
				if next != -1 {
					m.sessions[m.cursor], m.sessions[next] =
						m.sessions[next], m.sessions[m.cursor]
					m.cursor = next
				}
			}

		case "k":
			if m.mode == normalMode {
				// Move cursor up to previous visible session
				prev := m.findPrevVisibleSession(m.cursor)
				if prev != -1 {
					m.cursor = prev
				}
			} else {
				// Move item up - swap with previous visible session
				prev := m.findPrevVisibleSession(m.cursor)
				if prev != -1 {
					m.sessions[m.cursor], m.sessions[prev] =
						m.sessions[prev], m.sessions[m.cursor]
					m.cursor = prev
				}
			}

		case "h":
			// Switch to previous group
			if m.groupsData != nil && len(m.groupsData.Groups) > 1 {
				storage.PrevGroup(m.groupsData, m.wrapAround)
				// Update sessions to show current group
				m.sessions = m.groupsData.Groups[m.groupsData.CurrentGroup].Sessions
				m.cursor = 0 // Reset cursor
				m.ensureCursorOnVisibleSession()
				// Auto-save current group immediately so CLI commands see it
				storage.SaveGroupsData(m.groupsData)
			}

		case "l":
			// Switch to next group
			if m.groupsData != nil && len(m.groupsData.Groups) > 1 {
				storage.NextGroup(m.groupsData, m.wrapAround)
				// Update sessions to show current group
				m.sessions = m.groupsData.Groups[m.groupsData.CurrentGroup].Sessions
				m.cursor = 0 // Reset cursor
				m.ensureCursorOnVisibleSession()
				// Auto-save current group immediately so CLI commands see it
				storage.SaveGroupsData(m.groupsData)
			}

		case "n":
			// Create new group
			if m.groupsData != nil {
				m.mode = inputMode
				m.inputAction = createGroupAction
				m.textInput = textinput.New()
				m.textInput.Placeholder = "Enter group name"
				m.textInput.Focus()
			}

		case "r":
			// Rename current group
			if m.groupsData != nil {
				m.mode = inputMode
				m.inputAction = renameGroupAction
				m.textInput = textinput.New()
				m.textInput.Placeholder = "Enter new name"
				// Pre-fill with current name
				currentName := m.groupsData.Groups[m.groupsData.CurrentGroup].Name
				m.textInput.SetValue(currentName)
				m.textInput.Focus()
			}

		case "X":
			// Delete current group (with confirmation)
			if m.groupsData != nil && len(m.groupsData.Groups) > 1 {
				groupName := m.groupsData.Groups[m.groupsData.CurrentGroup].Name
				sessionCount := len(m.groupsData.Groups[m.groupsData.CurrentGroup].Sessions)
				m.confirmPrompt = fmt.Sprintf("Delete group '%s' (%d sessions)? (y/n)", groupName, sessionCount)
				m.mode = confirmMode
			}

		case "H":
			// Move current session to previous group
			if m.groupsData != nil && len(m.groupsData.Groups) > 1 && m.cursor < len(m.sessions) {
				if err := storage.MoveSessionToPrevGroup(m.groupsData, m.cursor, m.wrapAround); err == nil {
					// Update sessions after move
					m.sessions = m.groupsData.Groups[m.groupsData.CurrentGroup].Sessions
					// Adjust cursor if needed
					if m.cursor >= len(m.sessions) && len(m.sessions) > 0 {
						m.cursor = len(m.sessions) - 1
					}
				}
			}

		case "L":
			// Move current session to next group
			if m.groupsData != nil && len(m.groupsData.Groups) > 1 && m.cursor < len(m.sessions) {
				if err := storage.MoveSessionToNextGroup(m.groupsData, m.cursor, m.wrapAround); err == nil {
					// Update sessions after move
					m.sessions = m.groupsData.Groups[m.groupsData.CurrentGroup].Sessions
					// Adjust cursor if needed
					if m.cursor >= len(m.sessions) && len(m.sessions) > 0 {
						m.cursor = len(m.sessions) - 1
					}
				}
			}

		case "enter", "s":
			// Save and quit
			if m.groupsData != nil {
				// Update current group's sessions before saving
				m.groupsData.Groups[m.groupsData.CurrentGroup].Sessions = m.sessions
				// Save immediately in groups mode
				storage.SaveGroupsData(m.groupsData)
			} else if m.onSave != nil {
				// Legacy mode: use callback
				if err := m.onSave(m.sessions); err != nil {
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

	// Title with group navigation
	title := "✨ Rolo - Tmux Rolodex"
	if m.groupsData != nil {
		groupNav := getAllGroupNames(m.groupsData)
		if groupNav != "" {
			title = title + "\n" + groupNav
		}
	}
	s += titleStyle.Render(title) + "\n\n"

	// Mode indicator and help text
	if m.mode == confirmMode {
		s += helpStyle.Render(m.confirmPrompt) + "\n\n"
	} else if m.mode == inputMode {
		var prompt string
		if m.inputAction == createGroupAction {
			prompt = "Create new group:"
		} else if m.inputAction == renameGroupAction {
			prompt = "Rename group:"
		}
		s += helpStyle.Render(prompt) + "\n"
		s += m.textInput.View() + "\n"
		s += helpStyle.Render(keybindStyle.Render("enter")+" submit  "+keybindStyle.Render("esc")+" cancel") + "\n\n"
	} else if m.mode == moveMode {
		modeText := modeMoveStyle.Render("MOVE MODE")
		help := helpStyle.Render(
			keybindStyle.Render("j/k") + " move item  " +
				keybindStyle.Render("m") + " exit move mode",
		)
		s += modeText + " - " + help + "\n\n"
	} else {
		modeText := modeNormalStyle.Render("NORMAL")
		s += modeText

		if m.showKeymap {
			// Build help text with line wrapping
			helpLines := []string{}
			line := keybindStyle.Render("j/k") + " navigate  " +
				keybindStyle.Render("o") + " attach  " +
				keybindStyle.Render("d") + " delete  " +
				keybindStyle.Render("D") + " kill  " +
				keybindStyle.Render("t") + " toggle deleted  "
			helpLines = append(helpLines, line)

			line = keybindStyle.Render("u") + " update  " +
				keybindStyle.Render("p") + " repopulate  " +
				keybindStyle.Render("m") + " move  " +
				keybindStyle.Render("s/enter") + " save"

			if m.groupsData != nil {
				if len(m.groupsData.Groups) > 1 {
					line += "  " + keybindStyle.Render("h/l") + " switch"
				}
				helpLines = append(helpLines, line)

				line = ""
				if len(m.groupsData.Groups) > 1 {
					line = keybindStyle.Render("H/L") + " move session  "
				}
				line += keybindStyle.Render("n") + " new  " +
					keybindStyle.Render("r") + " rename"
				if len(m.groupsData.Groups) > 1 {
					line += "  " + keybindStyle.Render("X") + " delete"
				}
				helpLines = append(helpLines, line)
			} else {
				helpLines = append(helpLines, line)
			}

			line = keybindStyle.Render("?") + " toggle help"
			helpLines = append(helpLines, line)

			helpText := strings.Join(helpLines, "\n")
			help := helpStyle.Render(helpText)
			s += "\n" + help
		} else {
			help := helpStyle.Render(" " + keybindStyle.Render("?") + " for help")
			s += help
		}
		s += "\n\n"
	}

	// Session list (skip in input/confirm mode)
	if m.mode != inputMode && m.mode != confirmMode {
		for i, session := range m.sessions {
			// Skip deleted sessions if showDeleted is false
			if !m.showDeleted && session.Deleted {
				continue
			}

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
	}

	// Footer
	if m.mode != inputMode && m.mode != confirmMode {
		s += "\n" + helpStyle.Render("Press ") + keybindStyle.Render("q") + helpStyle.Render(", ") + keybindStyle.Render("esc") + helpStyle.Render(", or ") + keybindStyle.Render("ctrl+c") + helpStyle.Render(" to quit without saving")
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

	// Try to get current tmux session and position cursor on it
	cursor := 0
	currentSession, err := tmux.GetCurrentSession()
	if err == nil {
		// Find the index of current session in the list
		for i, session := range sessions {
			if session.Name == currentSession {
				cursor = i
				break
			}
		}
	}

	m := model{
		sessions:    sessions,
		cursor:      cursor,
		mode:        normalMode,
		onSave:      onSave,
		wrapAround:  config.WrapAround,
		groupsData:  nil, // Will be set when we refactor main.go
		showDeleted: config.ShowDeleted,
		config:      config,
	}

	// Ensure cursor starts on a visible session
	m.ensureCursorOnVisibleSession()

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	return nil
}

// RunWithGroups starts the interactive TUI with group support
func RunWithGroups(groupsData *storage.GroupsData, onSave func(*storage.GroupsData) error) error {
	// Load config to get wrap around setting
	config, err := storage.LoadConfig()
	if err != nil {
		config = &storage.Config{WrapAround: false}
	}

	// Get current group's sessions
	sessions := groupsData.Groups[groupsData.CurrentGroup].Sessions

	// Try to get current tmux session and position cursor on it
	cursor := 0
	currentSession, err := tmux.GetCurrentSession()
	if err == nil {
		// Find the index of current session in the list
		for i, session := range sessions {
			if session.Name == currentSession {
				cursor = i
				break
			}
		}
	}

	m := model{
		sessions:    sessions,
		cursor:      cursor,
		mode:        normalMode,
		onSave:      nil, // Old callback not used in groups mode
		wrapAround:  config.WrapAround,
		groupsData:  groupsData,
		showDeleted: config.ShowDeleted,
		config:      config,
	}

	// Ensure cursor starts on a visible session
	m.ensureCursorOnVisibleSession()

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	// Save groups data after TUI exits
	if onSave != nil {
		return onSave(groupsData)
	}

	return nil
}
