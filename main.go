package main

import (
	"fmt"
	"os"

	"rolo/storage"
	"rolo/tmux"
	"rolo/tui"
)

func saveOrder(sessions []string) error {
	return storage.SaveSessions(sessions)
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

	if err := storage.SaveSessions(sessions); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving sessions: %v\n", err)
		os.Exit(1)
	}

	configPath, _ := storage.GetConfigPath()
	fmt.Printf("Saved %d session(s) to %s:\n", len(sessions), configPath)
	for _, session := range sessions {
		fmt.Printf("  - %s\n", session)
	}
}

func findSessionIndex(sessions []string, target string) int {
	for i, session := range sessions {
		if session == target {
			return i
		}
	}
	return -1
}

func handleNext() {
	// Get current session
	currentSession, err := tmux.GetCurrentSession()
	if err != nil {
		os.Exit(1)
	}

	// Load ordered sessions
	sessions, err := storage.LoadSessions()
	if err != nil {
		os.Exit(1)
	}

	if len(sessions) == 0 {
		os.Exit(1)
	}

	// Find current session index
	currentIndex := findSessionIndex(sessions, currentSession)
	if currentIndex == -1 {
		os.Exit(1)
	}

	// Calculate next index (wrap around)
	nextIndex := (currentIndex + 1) % len(sessions)
	nextSession := sessions[nextIndex]

	// Switch to next session
	if err := tmux.SwitchToSession(nextSession); err != nil {
		os.Exit(1)
	}
}

func handlePrev() {
	// Get current session
	currentSession, err := tmux.GetCurrentSession()
	if err != nil {
		os.Exit(1)
	}

	// Load ordered sessions
	sessions, err := storage.LoadSessions()
	if err != nil {
		os.Exit(1)
	}

	if len(sessions) == 0 {
		os.Exit(1)
	}

	// Find current session index
	currentIndex := findSessionIndex(sessions, currentSession)
	if currentIndex == -1 {
		os.Exit(1)
	}

	// Calculate previous index (wrap around)
	prevIndex := (currentIndex - 1 + len(sessions)) % len(sessions)
	prevSession := sessions[prevIndex]

	// Switch to previous session
	if err := tmux.SwitchToSession(prevSession); err != nil {
		os.Exit(1)
	}
}

func runInteractiveMode() {
	// Load sessions from storage
	sessions, err := storage.LoadSessions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading sessions: %v\n", err)
		os.Exit(1)
	}

	// If no sessions exist, provide a helpful message
	if len(sessions) == 0 {
		sessions = []string{
			"No sessions found",
			"Run 'rolo populate' to fetch tmux sessions",
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
