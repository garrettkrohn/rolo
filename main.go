package main

import (
	"fmt"
	"os"

	"rolo/storage"
	"rolo/tmux"
	"rolo/tui"
)

func saveOrder(sessions []storage.SessionData) error {
	return storage.SaveSessionsData(sessions)
}

func showUsage() {
	fmt.Println("Usage:")
	fmt.Println("  rolo          - Launch interactive session reorder UI")
	fmt.Println("  rolo populate - Fetch active tmux sessions and save to config")
	fmt.Println("  rolo next     - Switch to next session in order")
	fmt.Println("  rolo prev     - Switch to previous session in order")
	fmt.Println("  rolo help     - Show this help message")
}

func handlePopulate() {
	sessions, err := tmux.GetActiveSessions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting tmux sessions: %v\n", err)
		os.Exit(1)
	}

	if len(sessions) == 0 {
		fmt.Println("No active tmux sessions found")
		return
	}

	// Convert to SessionData format (all non-deleted by default)
	sessionData := make([]storage.SessionData, len(sessions))
	for i, name := range sessions {
		sessionData[i] = storage.SessionData{Name: name, Deleted: false}
	}

	if err := storage.SaveSessionsData(sessionData); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving sessions: %v\n", err)
		os.Exit(1)
	}

	configPath, _ := storage.GetConfigJSONPath()
	fmt.Printf("Saved %d session(s) to %s:\n", len(sessions), configPath)
	for _, session := range sessions {
		fmt.Printf("  - %s\n", session)
	}
}

func findSessionIndex(sessions []storage.SessionData, target string) int {
	for i, session := range sessions {
		if session.Name == target {
			return i
		}
	}
	return -1
}

func findNextActiveSession(sessions []storage.SessionData, currentIndex int, wrapAround bool) int {
	if len(sessions) == 0 {
		return -1
	}
	
	// Determine the search range based on wrapAround setting
	maxIterations := len(sessions)
	if !wrapAround {
		// Only search from current position to end
		maxIterations = len(sessions) - currentIndex - 1
	}
	
	// Start from the next index
	for i := 1; i <= maxIterations; i++ {
		var nextIndex int
		if wrapAround {
			// Wrap around to the beginning
			nextIndex = (currentIndex + i) % len(sessions)
		} else {
			// Don't wrap around
			nextIndex = currentIndex + i
			if nextIndex >= len(sessions) {
				break
			}
		}
		
		if !sessions[nextIndex].Deleted {
			return nextIndex
		}
	}
	
	// No active session found
	return -1
}

func findPrevActiveSession(sessions []storage.SessionData, currentIndex int, wrapAround bool) int {
	if len(sessions) == 0 {
		return -1
	}
	
	// Determine the search range based on wrapAround setting
	maxIterations := len(sessions)
	if !wrapAround {
		// Only search from current position to beginning
		maxIterations = currentIndex
	}
	
	// Start from the previous index
	for i := 1; i <= maxIterations; i++ {
		var prevIndex int
		if wrapAround {
			// Wrap around to the end
			prevIndex = (currentIndex - i + len(sessions)) % len(sessions)
		} else {
			// Don't wrap around
			prevIndex = currentIndex - i
			if prevIndex < 0 {
				break
			}
		}
		
		if !sessions[prevIndex].Deleted {
			return prevIndex
		}
	}
	
	// No active session found
	return -1
}

func handleNext() {
	// Load config
	config, err := storage.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Get current session
	currentSession, err := tmux.GetCurrentSession()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current session: %v\n", err)
		os.Exit(1)
	}

	// Load ordered sessions
	sessions, err := storage.LoadSessionsData()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading sessions: %v\n", err)
		os.Exit(1)
	}

	if len(sessions) == 0 {
		fmt.Fprintf(os.Stderr, "No sessions configured. Run 'rolo populate' first.\n")
		os.Exit(1)
	}

	// Find current session index
	currentIndex := findSessionIndex(sessions, currentSession)
	if currentIndex == -1 {
		// Current session not in list, attach to first active session
		currentIndex = -1 // Start from beginning
	}

	// Try to find next active session, skipping ones that don't exist
	tried := 0
	maxAttempts := len(sessions)
	
	for tried < maxAttempts {
		// Find next active (non-deleted) session
		nextIndex := findNextActiveSession(sessions, currentIndex, config.WrapAround)
		if nextIndex == -1 {
			if !config.WrapAround {
				// Fail silently when at the end and not wrapping
				return
			} else {
				fmt.Fprintf(os.Stderr, "No active sessions available (all are deleted)\n")
				os.Exit(1)
			}
		}
		
		nextSession := sessions[nextIndex].Name
		
		// Try to switch to next session
		if err := tmux.SwitchToSession(nextSession); err != nil {
			// Log the error and mark session as deleted
			fmt.Fprintf(os.Stderr, "Warning: Session '%s' doesn't exist, skipping: %v\n", nextSession, err)
			sessions[nextIndex].Deleted = true
			
			// Save the updated state
			if saveErr := storage.SaveSessionsData(sessions); saveErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to save updated session state: %v\n", saveErr)
			}
			
			// Try the next one
			currentIndex = nextIndex
			tried++
			continue
		}
		
		// Success!
		return
	}
	
	// If we've tried all sessions and none worked
	fmt.Fprintf(os.Stderr, "Error: No valid sessions found\n")
	os.Exit(1)
}

func handlePrev() {
	// Load config
	config, err := storage.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Get current session
	currentSession, err := tmux.GetCurrentSession()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current session: %v\n", err)
		os.Exit(1)
	}

	// Load ordered sessions
	sessions, err := storage.LoadSessionsData()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading sessions: %v\n", err)
		os.Exit(1)
	}

	if len(sessions) == 0 {
		fmt.Fprintf(os.Stderr, "No sessions configured. Run 'rolo populate' first.\n")
		os.Exit(1)
	}

	// Find current session index
	currentIndex := findSessionIndex(sessions, currentSession)
	if currentIndex == -1 {
		// Current session not in list, attach to first active session
		currentIndex = -1 // Start from beginning
	}

	// Try to find previous active session, skipping ones that don't exist
	tried := 0
	maxAttempts := len(sessions)
	
	for tried < maxAttempts {
		// Find previous active (non-deleted) session
		prevIndex := findPrevActiveSession(sessions, currentIndex, config.WrapAround)
		if prevIndex == -1 {
			if !config.WrapAround {
				// Fail silently when at the beginning and not wrapping
				return
			} else {
				fmt.Fprintf(os.Stderr, "No active sessions available (all are deleted)\n")
				os.Exit(1)
			}
		}
		
		prevSession := sessions[prevIndex].Name
		
		// Try to switch to previous session
		if err := tmux.SwitchToSession(prevSession); err != nil {
			// Log the error and mark session as deleted
			fmt.Fprintf(os.Stderr, "Warning: Session '%s' doesn't exist, skipping: %v\n", prevSession, err)
			sessions[prevIndex].Deleted = true
			
			// Save the updated state
			if saveErr := storage.SaveSessionsData(sessions); saveErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to save updated session state: %v\n", saveErr)
			}
			
			// Try the next one
			currentIndex = prevIndex
			tried++
			continue
		}
		
		// Success!
		return
	}
	
	// If we've tried all sessions and none worked
	fmt.Fprintf(os.Stderr, "Error: No valid sessions found\n")
	os.Exit(1)
}

func runInteractiveMode() {
	// Load sessions from storage
	sessions, err := storage.LoadSessionsData()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading sessions: %v\n", err)
		os.Exit(1)
	}

	// If no sessions exist, provide a helpful message
	if len(sessions) == 0 {
		sessions = []storage.SessionData{
			{Name: "No sessions found", Deleted: false},
			{Name: "Run 'rolo populate' to fetch tmux sessions", Deleted: false},
		}
	}

	// Run the TUI
	if err := tui.Run(sessions, saveOrder); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	// Parse command line arguments
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "populate":
			handlePopulate()
			return
		case "next":
			handleNext()
			return
		case "prev", "previous":
			handlePrev()
			return
		case "help", "-h", "--help":
			showUsage()
			return
		default:
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
			showUsage()
			os.Exit(1)
		}
	}

	// No arguments - run interactive mode
	runInteractiveMode()
}
